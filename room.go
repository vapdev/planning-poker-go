package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type RoomRequest struct {
	UserUUID string `json:"userUUID"`
	RoomName string `json:"roomName"`
}

type JoinRoomRequest struct {
	UserUUID string `json:"UserUUID"`
	RoomUUID string `json:"RoomUUID"`
}

func createRoom(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RoomRequest
		if err := json.NewDecoder(r.Body).Decode(&req); handleError(w, err) {
			return
		}

		roomUUID, userUUID, roomName, err := createRoomInDB(database, req.UserUUID, req.RoomName)
		if handleError(w, err) {
			return
		}

		userID, _ := getUserIDFromUUID(database, userUUID)
		roomID, _ := getRoomIDFromUUID(database, roomUUID)

		games[roomUUID] = &Game{
			Players: []*Player{},
			admin:   int(userID),
			roomID:  roomID,
			name:    roomName,
		}
		sendGameState(games[roomUUID])

		sendResponse(w, map[string]interface{}{
			"roomUUID": roomUUID,
			"userUUID": userUUID,
			"roomName": roomName,
		})
	}
}

func sendPlayerLeftMessage(game *Game, userID int) {
	handleLeaveRoom(game, userID)
	sendGameState(game)
}

func leaveRoom(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		log.Printf("Params: %v", params)
		roomUUID, roomExists := params["roomUUID"].(string)
		userUUID, userExists := params["userUUID"].(string)

		roomID, _ := getRoomIDFromUUID(database, roomUUID)
		userID, _ := getUserIDFromUUID(database, userUUID)

		if !roomExists || !userExists {
			http.Error(w, "Room ID or User ID not provided", http.StatusBadRequest)
			return
		}

		// Check if the user is the admin
		var adminID int
		err = database.QueryRow("SELECT admin FROM rooms WHERE id = $1", int(roomID)).Scan(&adminID)
		if err != nil {
			http.Error(w, "Failed to retrieve admin from database", http.StatusInternalServerError)
			return
		}

		if adminID == int(userID) {
			// If the user is the admin, set another player as the admin
			_, err = database.Exec("UPDATE rooms SET admin = (SELECT user_id FROM room_users WHERE room_id = $1 LIMIT 1) WHERE id = $2", int(roomID), int(roomID))
			if err != nil {
				http.Error(w, "Failed to update admin in database", http.StatusInternalServerError)
				return
			}
		}

		// Delete the record from the database
		_, err = database.Exec("DELETE FROM room_users WHERE room_id = $1 AND user_id = $2", int(roomID), int(userID))
		if err != nil {
			http.Error(w, "Failed to delete record from database", http.StatusInternalServerError)
			return
		}

		game, gameExists := games[roomUUID]
		if gameExists {
			sendPlayerLeftMessage(game, int(userID))
		}
	}
}

func joinRoom(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req JoinRoomRequest
		if err := json.NewDecoder(r.Body).Decode(&req); handleError(w, err) {
			return
		}

		roomUUID, userUUID, err := addUserToRoom(database, req.RoomUUID, req.UserUUID)

		if handleError(w, err) {
			return
		}

		if handleError(w, err) {
			return
		}

		sendGameState(games[roomUUID])

		sendResponse(w, map[string]interface{}{
			"roomUUID": roomUUID,
			"userUUID": userUUID,
		})
	}
}

func resetVotes(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			RoomUUID string `json:"roomUUID"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); handleError(w, err) {
			return
		}

		roomID, err := getRoomIDFromUUID(database, req.RoomUUID)
		if handleError(w, err) {
			return
		}

		_, err = database.Exec("DELETE FROM votes WHERE room_id = $1", roomID)
		if handleError(w, err) {
			return
		}

		_, err = database.Exec("UPDATE rooms SET showCards = false WHERE id = $1", roomID)
		if handleError(w, err) {
			return
		}

		game, exists := games[req.RoomUUID]
		if exists {
			game.showCards = false
			for _, player := range game.Players {
				player.Voted = false
				player.Vote = nil // Set player.Vote to nil instead of 0
			}
			sendGameState(game)
		}
	}
}

func autoShowCards(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Auto show cards")

		var req struct {
			RoomUUID string `json:"roomUUID"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); handleError(w, err) {
			log.Println("Error decoding request body")
			return
		}

		log.Println("uuid: ", req.RoomUUID)
		RoomID, _ := getRoomIDFromUUID(database, req.RoomUUID)
		log.Printf("Room ID: %d", RoomID)

		var currentAutoShowState bool
		err := database.QueryRow("SELECT autoShowCards FROM rooms WHERE id = $1", RoomID).Scan(&currentAutoShowState)
		if handleError(w, err) {
			return
		}

		newAutoShowState := !currentAutoShowState
		_, err = database.Exec("UPDATE rooms SET autoShowCards = $1 WHERE id = $2", newAutoShowState, RoomID)
		if handleError(w, err) {
			return
		}
		log.Println("Auto show cards end")

		game, exists := games[req.RoomUUID]
		if exists {
			game.autoShowCards = newAutoShowState
			if !newAutoShowState {
				game.showCards = false
				_, err = database.Exec("UPDATE rooms SET showCards = $1 WHERE id = $2", false, RoomID)
				if handleError(w, err) {
					return
				}
			}
			sendGameState(game)
		}
	}
}

func showCards(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			RoomUUID string `json:"roomUUID"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); handleError(w, err) {
			return
		}

		RoomID, err := getRoomIDFromUUID(database, req.RoomUUID)
		log.Printf("Room ID: %d", RoomID)
		log.Printf("Room UUID: %s", req.RoomUUID)
		if err != nil {
			http.Error(w, "Failed to get room ID from UUID", http.StatusInternalServerError)
			return
		}

		var currentShowState bool
		err = database.QueryRow("SELECT showCards FROM rooms WHERE id = $1", RoomID).Scan(&currentShowState)
		if handleError(w, err) {
			return
		}

		newShowState := !currentShowState
		_, err = database.Exec("UPDATE rooms SET showCards = $1 WHERE id = $2", newShowState, RoomID)
		if handleError(w, err) {
			return
		}

		game, exists := games[req.RoomUUID]
		if exists {
			game.showCards = newShowState
			sendGameState(game)
		}
	}
}

func vote(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			UserID int `json:"userID"`
			RoomID int `json:"roomID"`
			Vote   int `json:"vote"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); handleError(w, err) {
			return
		}

		if err := castVote(database, req.RoomID, req.UserID, req.Vote); handleError(w, err) {
			return
		}

		fmt.Fprintf(w, "User %d voted %d in room %d\n", req.UserID, req.Vote, req.RoomID)
	}
}

func generateUuid() string {
	return uuid.New().String()
}

func createRoomInDB(database *sql.DB, userUUID string, roomName string) (string, string, string, error) {
	log.Printf("Creating room with name %s", roomName)
	log.Printf("User UUID: %s", userUUID)
	tx, err := database.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return "", "", "", err
	}

	roomUUID, err := generateRoomUUID(database)
	if err != nil {
		log.Printf("Error generating room UUID: %v", err)
		return "", "", "", err
	}

	if userUUID == "" {
		userUUID = generateUuid()
	}
	userID, err := getUserIDFromUUID(database, userUUID)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No user found with UUID %s, generating new UUID", userUUID)
			userUUID = generateUuid()
			row := tx.QueryRow("INSERT INTO users (name, uuid) VALUES ('Guest', $1) RETURNING id", userUUID)
			err = row.Scan(&userID)
			if err != nil {
				log.Printf("Error inserting user: %v", err)
				return "", "", "", err
			}
		} else {
			log.Printf("Error getting user ID from UUID: %v", err)
			return "", "", "", err
		}
	}

	var roomID int64
	statement := `INSERT INTO rooms (uuid, admin, name) VALUES ($1, $2, $3) RETURNING id`
	err = tx.QueryRow(statement, roomUUID, userID, roomName).Scan(&roomID)
	if err != nil {
		log.Printf("Error inserting room: %v", err)
		return "", "", "", err
	}

	statement = "INSERT INTO room_users (room_id, user_id) VALUES ($1, $2)"
	_, err = tx.Exec(statement, roomID, userID)
	if err != nil {
		log.Printf("Error inserting room_user: %v", err)
		return "", "", "", err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return "", "", "", err
	}
	return roomUUID, userUUID, roomName, nil
}

func addUserToRoom(database *sql.DB, roomUUID string, userUUID string) (string, string, error) {
	var userID int64
	var err error

	// Try to get the user ID from the provided UUID.
	if userUUID != "" {
		userID, err = getUserIDFromUUID(database, userUUID)
		if err != nil && err != sql.ErrNoRows {
			// If there's an error other than sql.ErrNoRows, return it.
			return "", "", err
		}
	}

	// If no userUUID was provided, or no user was found for the provided UUID,
	// generate a new UUID and create a new user.
	if userUUID == "" || err == sql.ErrNoRows {
		userUUID = generateUuid()
		var userID int64
		err := database.QueryRow("INSERT INTO users (name, uuid) VALUES ('Guest', $1) RETURNING id", userUUID).Scan(&userID)
		if err != nil {
			return "", "", err
		}
	}

	roomID, err := getRoomIDFromUUID(database, roomUUID)
	if err != nil {
		return "", "", err
	}

	// Check if user is already in the room
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM room_users WHERE room_id = $1 AND user_id = $2", roomID, userID).Scan(&count)
	if err != nil {
		return "", "", err
	}
	if count > 0 {
		log.Printf("User %d is already in room %d", userID, roomID)
		return roomUUID, userUUID, nil
	}

	statement, err := database.Prepare("INSERT INTO room_users (room_id, user_id) VALUES ($1, $2)")
	if err != nil {
		return "", "", err
	}
	_, err = statement.Exec(roomID, userID)
	if err != nil {
		return "", "", err
	}

	return roomUUID, userUUID, nil
}
func castVote(database *sql.DB, roomID, userID, vote int) error {
	// Check if a vote from the user already exists
	var existingVote int
	err := database.QueryRow("SELECT vote FROM votes WHERE room_id = $1 AND user_id = $2", roomID, userID).Scan(&existingVote)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		// No existing vote, insert new vote
		statement, err := database.Prepare("INSERT INTO votes (room_id, user_id, vote) VALUES ($1, $2, $3)")
		if err != nil {
			return err
		}
		_, err = statement.Exec(roomID, userID, vote)
		if err != nil {
			return err
		}
	} else {
		// Existing vote
		if existingVote == vote {
			// Same vote, remove it
			statement, err := database.Prepare("DELETE FROM votes WHERE room_id = $1 AND user_id = $2")
			if err != nil {
				return err
			}
			_, err = statement.Exec(roomID, userID)
			if err != nil {
				return err
			}
		} else {
			// Different vote, update it
			statement, err := database.Prepare("UPDATE votes SET vote = $1 WHERE room_id = $2 AND user_id = $3")
			if err != nil {
				return err
			}
			_, err = statement.Exec(vote, roomID, userID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
