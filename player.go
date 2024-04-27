package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

func changeName(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			UserUUID string `json:"userUUID"`
			RoomUUID string `json:"roomUUID"`
			Name     string `json:"name"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		UserID, err := getUserIDFromUUID(database, req.UserUUID)
		if err != nil {
			log.Printf("Error getting user ID: %v", err)
			return
		}

		statement, err := database.Prepare("UPDATE users SET name = $1 WHERE id = $2")
		if err != nil {
			log.Printf("Error preparing statement: %v", err)
			return
		}

		_, err = statement.Exec(req.Name, UserID)
		if err != nil {
			log.Printf("Error executing statement: %v", err)
			return
		}
		log.Printf("User %d changed name to %s", UserID, req.Name)
		game, exists := games[req.RoomUUID]
		log.Printf("%v", games)
		log.Printf("players of game %v", game.Players)

		if exists {
			for i := range game.Players {
				if game.Players[i].ID == int(UserID) {
					log.Printf("Changing name of player %d to %s", game.Players[i].ID, req.Name)
					game.Players[i].Name = req.Name
					break
				}
			}
			log.Printf("Sending game state to room %s", req.RoomUUID)
		}
		sendGameState(game)
	}
}
