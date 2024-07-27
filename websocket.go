package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
    upgrader = websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
        CheckOrigin:     func(r *http.Request) bool { return true },
    }
	db *sql.DB
    gamesMu  sync.Mutex               // Mutex to protect access to the games map
)

func getDB() *sql.DB {
    if db == nil {
        var err error
        db, err = sql.Open("pgx", os.Getenv("DATABASE_URL"))
        if err != nil {
            log.Fatal(err)
        }
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
		handleNewPlayer(msg, game, int(userID), userUUID, ws)
	case "newAdmin":
		handleNewAdmin(msg, game, int(userID), userUUID, ws)
	case "playerLeft":
		handleLeaveRoom(game, int(userID))
	}
	game.lastActive = time.Now()
	sendGameState(game)
}

func sendGameState(game *Game) {
	// Check if the game object is nil
	if game == nil {
		log.Println("Game object is nil, cannot send game state")
		return
	}

	// Ensure the Players slice is initialized
	if game.Players == nil {
		log.Println("Players slice is nil, initializing to empty slice")
		game.Players = []*Player{}
	}

	// Update showCards based on votes if autoShowCards is enabled
	if game.autoShowCards {
		allVoted := true
		for _, player := range game.Players {
			if player == nil {
				log.Println("Player in Players slice is nil")
				continue
			}
			if !player.Voted {
				allVoted = false
				break
			}
		}
		if allVoted {
			game.showCards = true
		}
	}

	// Prepare the message to send to players
	msg := map[string]interface{}{
		"type":          "gameState",
		"players":       game.Players,
		"showCards":     game.showCards,
		"autoShowCards": game.autoShowCards,
		"roomUUID":      game.roomUUID,
		"name":          game.name,
		"admin":         game.admin,
	}

	// Send the game state to each player
	for _, player := range game.Players {
		if player == nil {
			log.Println("Player is nil, skipping")
			continue
		}
		if player.ws == nil {
			log.Printf("WebSocket connection for player %d is nil, skipping", player.ID)
			continue
		}

		err := player.ws.WriteJSON(msg)
		if err != nil {
			log.Printf("Error writing JSON to WebSocket for player %d: %v", player.ID, err)
		}
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	roomUUID, roomExists := params["roomUUID"]
	userUUID, userExists := params["userUUID"]


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

    done := make(chan struct{})
    defer close(done)

    ws.SetPongHandler(func(appData string) error {
        log.Println("Received pong")
        return nil 
    })

    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()
        for {
            select {
            case <-done:
                return
            case <-ticker.C:
                if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
                    log.Println("Error sending ping:", err)
                    return
                }
            }
        }
    }()

	gamesMu.Lock()
	game, gameExists := games[roomUUID]
	if !gameExists {
		db := getDB()
		game, err = fetchGameFromDB(db, roomUUID)
		if err != nil {
			log.Printf("Error fetching game from database: %v", err)
			gamesMu.Unlock()
			return
		}
		games[roomUUID] = game
	}
	gamesMu.Unlock()

	// Check if the user already exists in the game's players
	for _, player := range game.Players {
		if player.UUID == userUUID {
			log.Printf("User %s already exists in the game, replacing WebSocket connection", userUUID)
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

		handleMessage(msg, game, userUUID, ws)
	}
}
