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

		statement, err := database.Prepare("UPDATE users SET name = ? WHERE id = ?")
		if err != nil {
			log.Printf("Error preparing statement: %v", err)
			return
		}

		_, err = statement.Exec(req.Name, UserID)
		if err != nil {
			log.Printf("Error executing statement: %v", err)
			return
		}

		game, exists := games[req.RoomUUID]
		if exists {
			for i := range game.Players {
				if game.Players[i].ID == int(UserID) {
					game.Players[i].Name = req.Name
					break
				}
			}
			sendGameState(game)
		}
	}
}
