package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var (
	games         = make(map[string]*Game)
	roomExpiry    = 30 * time.Minute 
	cleanupTicker = time.NewTicker(5 * time.Minute)
)

func cleanupExpiredRooms() {
	for {
		select {
		case <-cleanupTicker.C:
			now := time.Now()
			for roomUUID, game := range games {
				if now.Sub(game.lastActive) > roomExpiry {
					log.Printf("Cleaning up expired room: %s", roomUUID)
					delete(games, roomUUID)
				}
			}
		}
	}
}

func main() {
	database := setupDatabase()

	log.Println("Setting up routes...")
	r := mux.NewRouter()
	r.HandleFunc("/ws/{roomUUID}/{userUUID}", enableCors(handleConnections))
	r.HandleFunc("/createRoom", enableCors(createRoom(database)))
	r.HandleFunc("/joinRoom", enableCors(joinRoom(database)))
	r.HandleFunc("/leaveRoom", enableCors(leaveRoom(database)))
	r.HandleFunc("/showCards", enableCors(showCards(database)))
	r.HandleFunc("/autoShowCards", enableCors(autoShowCards(database)))
	r.HandleFunc("/resetVotes", enableCors(resetVotes(database)))
	r.HandleFunc("/changeName", enableCors(changeName(database)))
	r.HandleFunc("/changeRoomName", enableCors(changeRoomName(database)))
	r.HandleFunc("/kickPlayer", enableCors(kickPlayer(database)))

	go cleanupExpiredRooms()

	log.Println("Starting server on " + os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), r))
}
