package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type RoomRequest struct {
	UserID int `json:"userID"`
}

type JoinRoomRequest struct {
	UserID int `json:"roomID"`
	RoomID int `json:"userID"`
}

func createRoom(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RoomRequest
		if err := json.NewDecoder(r.Body).Decode(&req); handleError(w, err) {
			return
		}

		roomID, userID, err := createRoomInDB(database, req.UserID)
		if handleError(w, err) {
			return
		}

		games[strconv.FormatInt(roomID, 10)] = &Game{
			Players: []*Player{},
			admin:   userID,
			roomID:  int(roomID),
		}

		sendResponse(w, map[string]interface{}{
			"roomID": int(roomID),
			"admin":  userID,
			"userID": userID,
		})
	}
}

func sendPlayerLeftMessage(game *Game, userID int) {
	msg := map[string]interface{}{
		"type":   "playerLeft",
		"userID": userID,
	}
	handleLeaveRoom(msg, game, userID)
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
		roomID, roomExists := params["roomID"].(float64)
		userID, userExists := params["userID"].(float64)

		if !roomExists || !userExists {
			http.Error(w, "Room ID or User ID not provided", http.StatusBadRequest)
			return
		}

		// Check if the user is the admin
		var adminID int
		err = database.QueryRow("SELECT admin FROM rooms WHERE id = ?", int(roomID)).Scan(&adminID)
		if err != nil {
			http.Error(w, "Failed to retrieve admin from database", http.StatusInternalServerError)
			return
		}

		if adminID == int(userID) {
			// If the user is the admin, set another player as the admin
			_, err = database.Exec("UPDATE rooms SET admin = (SELECT user_id FROM room_users WHERE room_id = ? LIMIT 1) WHERE id = ?", int(roomID), int(roomID))
			if err != nil {
				http.Error(w, "Failed to update admin in database", http.StatusInternalServerError)
				return
			}
		}

		// Delete the record from the database
		_, err = database.Exec("DELETE FROM room_users WHERE room_id = ? AND user_id = ?", int(roomID), int(userID))
		if err != nil {
			http.Error(w, "Failed to delete record from database", http.StatusInternalServerError)
			return
		}

		game, gameExists := games[strconv.Itoa(int(roomID))]
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

		roomID, userID, err := addUserToRoom(database, req.RoomID, req.UserID)

		if handleError(w, err) {
			return
		}

		roomIDStr := strconv.Itoa(roomID)
		if _, exists := games[roomIDStr]; !exists {
			games[roomIDStr] = &Game{
				Players: []*Player{},
				// Add other fields as necessary
			}
		}
		if handleError(w, err) {
			return
		}

		sendResponse(w, map[string]interface{}{
			"roomID": roomID,
			"userID": userID,
		})
	}
}

func resetVotes(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			RoomID int `json:"roomID"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); handleError(w, err) {
			return
		}

		_, err := database.Exec("DELETE FROM votes WHERE room_id = ?", req.RoomID)
		if handleError(w, err) {
			return
		}

		_, err = database.Exec("UPDATE rooms SET showCards = 0 WHERE id = ?", req.RoomID)
		if handleError(w, err) {
			return
		}

		game, exists := games[strconv.Itoa(req.RoomID)]
		if exists {
			game.showCards = false
			for _, player := range game.Players {
				player.Voted = false
				player.Vote = 0
			}
			sendGameState(game)
		}
	}
}

func autoShowCards(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			RoomID int `json:"roomID"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); handleError(w, err) {
			return
		}

		var currentAutoShowState bool
		err := database.QueryRow("SELECT autoShowCards FROM rooms WHERE id = ?", req.RoomID).Scan(&currentAutoShowState)
		if handleError(w, err) {
			return
		}

		newAutoShowState := !currentAutoShowState
		_, err = database.Exec("UPDATE rooms SET autoShowCards = ? WHERE id = ?", newAutoShowState, req.RoomID)
		if handleError(w, err) {
			return
		}

		game, exists := games[strconv.Itoa(req.RoomID)]
		if exists {
			game.autoShowCards = newAutoShowState
			if !newAutoShowState {
				game.showCards = false
				_, err = database.Exec("UPDATE rooms SET showCards = ? WHERE id = ?", false, req.RoomID)
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
			RoomID int `json:"roomID"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); handleError(w, err) {
			return
		}

		var currentShowState bool
		err := database.QueryRow("SELECT showCards FROM rooms WHERE id = ?", req.RoomID).Scan(&currentShowState)
		if handleError(w, err) {
			return
		}

		newShowState := !currentShowState
		_, err = database.Exec("UPDATE rooms SET showCards = ? WHERE id = ?", newShowState, req.RoomID)
		if handleError(w, err) {
			return
		}

		game, exists := games[strconv.Itoa(req.RoomID)]
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

func createRoomInDB(database *sql.DB, userID int) (int64, int, error) {
	tx, err := database.Begin()
	if err != nil {
		return 0, 0, err
	}

	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", userID).Scan(&count)
	if err != nil {
		return 0, 0, err
	}
	if count == 0 {
		_, err = tx.Exec("INSERT INTO users (id, name) VALUES (?, 'Admin')", userID)
		if err != nil {
			return 0, 0, err
		}
	}

	statement, err := tx.Prepare("INSERT INTO rooms (slug, admin) VALUES (?, ?)")
	if err != nil {
		return 0, 0, err
	}
	res, err := statement.Exec(generateRandomString(5), userID)
	if err != nil {
		return 0, 0, err
	}

	roomID, err := res.LastInsertId()
	if err != nil {
		return 0, 0, err
	}

	statement, err = tx.Prepare("INSERT INTO room_users (room_id, user_id) VALUES (?, ?)")
	if err != nil {
		return 0, 0, err
	}
	_, err = statement.Exec(roomID, userID)
	if err != nil {
		return 0, 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, 0, err
	}

	return roomID, userID, nil
}

func addUserToRoom(database *sql.DB, userID int, roomID int) (int, int, error) {
	if userID == 0 {
		log.Println("User not found, creating new user")
		err := database.QueryRow("INSERT INTO users (name) VALUES ('JoinUser') RETURNING id").Scan(&userID)
		if err != nil {
			return 0, 0, err
		}
	}
	log.Println(userID)

	// Check if user is already in the room
	var count int
	err := database.QueryRow("SELECT COUNT(*) FROM room_users WHERE room_id = ? AND user_id = ?", roomID, userID).Scan(&count)
	if err != nil {
		return 0, 0, err
	}
	if count > 0 {
		log.Printf("User %d is already in room %d", userID, roomID)
		return roomID, userID, nil
	}

	statement, err := database.Prepare("INSERT INTO room_users (room_id, user_id) VALUES (?, ?)")
	if err != nil {
		return 0, 0, err
	}
	_, err = statement.Exec(roomID, userID)
	if err != nil {
		return 0, 0, err
	}

	return roomID, userID, nil
}

func castVote(database *sql.DB, roomID, userID, vote int) error {
	// Check if a vote from the user already exists
	var existingVote int
	err := database.QueryRow("SELECT vote FROM votes WHERE room_id = ? AND user_id = ?", roomID, userID).Scan(&existingVote)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		// No existing vote, insert new vote
		statement, err := database.Prepare("INSERT INTO votes (room_id, user_id, vote) VALUES (?, ?, ?)")
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
			statement, err := database.Prepare("DELETE FROM votes WHERE room_id = ? AND user_id = ?")
			if err != nil {
				return err
			}
			_, err = statement.Exec(roomID, userID)
			if err != nil {
				return err
			}
		} else {
			// Different vote, update it
			statement, err := database.Prepare("UPDATE votes SET vote = ? WHERE room_id = ? AND user_id = ?")
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
