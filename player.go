package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func changeName(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			UserID int    `json:"userID"`
			RoomID int    `json:"roomID"`
			Name   string `json:"name"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Printf("Error decoding JSON: %v", err)
			return
		}

		statement, err := database.Prepare("UPDATE users SET name = ? WHERE id = ?")
		if err != nil {
			log.Printf("Error preparing statement: %v", err)
			return
		}

		_, err = statement.Exec(req.Name, req.UserID)
		if err != nil {
			log.Printf("Error executing statement: %v", err)
			return
		}

		game, exists := games[strconv.Itoa(req.RoomID)]
		if exists {
			for i := range game.Players {
				if game.Players[i].ID == req.UserID {
					game.Players[i].Name = req.Name
					break
				}
			}
			sendGameState(game)
		}

		log.Printf("User %d changed their name to %s", req.UserID, req.Name)
	}
}
