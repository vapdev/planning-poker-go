package main

import (
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	db      *sql.DB
	gamesMu sync.Mutex // Mutex to protect access to the games map
)

func handleMessage(msg map[string]interface{}, game *Game, userUUID string, ws *websocket.Conn, db *sql.DB) {
	userID, err := getUserIDFromUUID(db, userUUID)
	if err != nil {
		log.Printf("Error getting user ID from UUID: %v", err)
	}
	switch msg["type"] {
	case "vote":
		handleVote(msg, game, int(userID), db)
		sendGameState(game, nil) // Enviar estado do jogo sem emojis
	case "newPlayer":
		handleNewPlayer(msg, game, int(userID), userUUID, ws)
		sendGameState(game, nil) // Enviar estado do jogo sem emojis
	case "newAdmin":
		handleNewAdmin(msg, game, int(userID), userUUID, ws)
		sendGameState(game, nil) // Enviar estado do jogo sem emojis
	case "playerLeft":
		handleLeaveRoom(game, int(userID))
		sendGameState(game, nil) // Enviar estado do jogo sem emojis
	case "emoji":
		handleEmoji(msg, game, int(userID)) // A função `handleEmoji` já chama `sendGameState` com emojis
	default:
		sendGameState(game, nil) // Enviar estado do jogo sem emojis para outros tipos de mensagem
	}
	game.lastActive = time.Now()
}

func sendGameState(game *Game, emojis ...[]EmojiMessage) {
	// Check if emojis is provided, if not default to nil
	var emojiMessages []EmojiMessage
	if len(emojis) > 0 {
		emojiMessages = emojis[0]
	} else {
		emojiMessages = nil
	}

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
		allVoted := len(game.Players) > 0
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
		"emojis":        emojiMessages, // Include the emojis in the game state
		"deck":          game.deck,
	}

	// Send the game state to each player
	for _, player := range game.Players {
		if player == nil {
			log.Println("Player is nil, skipping")
			continue
		}

		// Try to send through all connections
		for i, conn := range player.connections {
			if conn == nil {
				continue
			}

			err := conn.WriteJSON(msg)
			if err != nil {
				log.Printf("Error in loop, writing JSON to WebSocket %d for player %d: %v", i, player.ID, err)
				// Remove failed connection
				player.connections[i] = nil
			}
		}

		// Clean up nil connections
		activeConns := make([]*websocket.Conn, 0)
		for _, conn := range player.connections {
			if conn != nil {
				activeConns = append(activeConns, conn)
			}
		}
		player.connections = activeConns
	}
}

func checkIfUserHasActiveConnections(game *Game, userID int) bool {
	for _, player := range game.Players {
		log.Printf("Checking player %d", player.ID)
		if player.ID == userID {
			activeConns := make([]*websocket.Conn, 0)
			log.Printf("Player %d has %d connections", player.ID, len(player.connections))
			for _, conn := range player.connections {
				log.Printf("Checking connection from player %d", player.ID)
				if conn != nil {
					log.Printf("Connection from player %d is not nil", player.ID)
					if err := conn.WriteMessage(websocket.PingMessage, nil); err == nil {
						log.Printf("Connection from player %d is active", player.ID)
						activeConns = append(activeConns, conn)
					}
				}
			}
			player.connections = activeConns
			return len(player.connections) > 0
		}
	}
	return false
}

func handleConnections(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			game, err = fetchGameFromDB(db, roomUUID)
			if err != nil {
				log.Printf("Error fetching game from database: %v", err)
				gamesMu.Unlock()
				return
			}
			games[roomUUID] = game
		}
		gamesMu.Unlock()

		for _, player := range game.Players {
			if player.UUID == userUUID {
				player.connections = append(player.connections, ws)
				break
			}
		}

		for {
			var msg map[string]interface{}
			err := ws.ReadJSON(&msg)
			//if close sent by client
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				userID, _ := getUserIDFromUUID(db, userUUID)
				//remove the connection from the player
				for _, player := range game.Players {
					if player.ID == userID {
						for i, conn := range player.connections {
							if conn == ws {
								player.connections = append(player.connections[:i], player.connections[i+1:]...)
								break
							}
						}
						break
					}
				}
				if !checkIfUserHasActiveConnections(game, userID) {
					log.Printf("User %d has no active connections, removing from room", userID)
					handleLeaveRoom(game, userID)
					sendGameState(game, nil)
				}
				break
			}

			if err != nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Printf("Error reading JSON from WebSocket: %v", err)
				break
			}

			handleMessage(msg, game, userUUID, ws, db)
		}
	}
}
