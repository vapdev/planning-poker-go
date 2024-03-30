package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var games = make(map[string]*Game)

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
	r.HandleFunc("/showCards", enableCors(showCards(database)))
	r.HandleFunc("/autoShowCards", enableCors(autoShowCards(database)))
	r.HandleFunc("/resetVotes", enableCors(resetVotes(database)))
	r.HandleFunc("/changeName", enableCors(changeName(database)))

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
