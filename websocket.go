package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func getDB() *sql.DB {
	db, err := sql.Open("sqlite3", "planningpoker.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func handleMessage(msg map[string]interface{}, game *Game, userUUID string, ws *websocket.Conn) {
	db := getDB()
	userID, err := getUserIDFromUUID(db, userUUID)
	if err != nil {
		log.Printf("Error getting user ID from UUID: %v", err)
	}
	switch msg["type"] {
	case "vote":
		handleVote(msg, game, int(userID))
	case "newPlayer":
		handleNewPlayer(msg, game, int(userID), ws)
	case "newAdmin":
		handleNewAdmin(msg, game, int(userID), ws)
	case "playerLeft":
		handleLeaveRoom(game, int(userID))
	}
	sendGameState(game)
}

func sendGameState(game *Game) {
	if game.autoShowCards {
		allVoted := true
		for _, player := range game.Players {
			if !player.Voted {
				allVoted = false
				break
			}
		}
		if allVoted {
			game.showCards = true
		}
	}

	for _, player := range game.Players {
		msg := map[string]interface{}{
			"type":          "gameState",
			"players":       game.Players,
			"showCards":     game.showCards,
			"autoShowCards": game.autoShowCards,
			"roomID":        game.roomID,
			"admin":         game.admin,
		}
		if player.ws != nil {
			err := player.ws.WriteJSON(msg)
			if err != nil {
				log.Printf("Error writing JSON to WebSocket: %v", err)
				break
			}
		}
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	roomUUID, roomExists := params["roomUUID"]
	userUUIDStr, userExists := params["userUUID"]

	log.Println("roomUUID: ", roomUUID)
	log.Println("userUUIDStr: ", userUUIDStr)

	if !roomExists || !userExists {
		log.Println("Room ID or User UUID not provided")
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer ws.Close()

	game, gameExists := games[roomUUID]
	if !gameExists {
		log.Printf("Game not found: %s", roomUUID)
		return
	}

	// Check if the user already exists in the game's players
	for _, player := range game.Players {
		if player.UUID == userUUIDStr {
			log.Printf("User %s already exists in the game, replacing WebSocket connection", userUUIDStr)
			player.ws = ws
		}
	}

	for {
		var msg map[string]interface{}
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading JSON from WebSocket: %v", err)
			break
		}

		handleMessage(msg, game, userUUIDStr, ws)
	}
}
