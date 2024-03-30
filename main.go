package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

type Player struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
	Voted bool   `json:"voted"`
	Vote  int    `json:"vote"`
	Admin bool   `json:"admin"`
	ws    *websocket.Conn
}

type Game struct {
	Players []*Player
	admin   int
}

var games = make(map[string]*Game)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	roomID, roomExists := params["roomID"]
	userIDStr, userExists := params["userID"]

	if !roomExists || !userExists {
		log.Println("Room ID or User ID not provided")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("User ID is not an integer: %v", err)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer ws.Close()

	game, gameExists := games[roomID]
	if !gameExists {
		log.Printf("Game not found: %s", roomID)
		return
	}

	// Check if the user already exists in the game's players
	for _, player := range game.Players {
		if player.ID == userID {
			log.Printf("User %d already exists in the game, replacing WebSocket connection", userID)
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

		handleMessage(msg, game, userID, ws)
	}
}
func handleMessage(msg map[string]interface{}, game *Game, userID int, ws *websocket.Conn) {
	switch msg["type"] {
	case "vote":
		handleVote(msg, game, userID)
	case "newPlayer":
		handleNewPlayer(msg, game, userID, ws)
	case "newAdmin":
		handleNewAdmin(msg, game, userID, ws)
	case "playerLeft":
		handleLeaveRoom(msg, game, userID)
	}
	sendGameState(game)
}

func sendGameState(game *Game) {
	for _, player := range game.Players {
		msg := map[string]interface{}{
			"type":    "gameState",
			"players": game.Players,
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

func main() {
	database, err := sql.Open("sqlite3", "planningpoker.db")
	if err != nil {
		log.Fatal(err)
	}
	createTables(database)

	log.Println("Setting up routes...")
	r := mux.NewRouter()
	r.HandleFunc("/ws/{roomID}/{userID}", enableCors(handleConnections))
	r.HandleFunc("/createRoom", enableCors(createRoom(database)))
	r.HandleFunc("/joinRoom", enableCors(joinRoom(database)))
	r.HandleFunc("/leaveRoom", enableCors(leaveRoom(database)))
	r.HandleFunc("/vote", enableCors(vote(database)))

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
