package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var games = make(map[string]*Game)

func main() {
	database := setupDatabase()

	log.Println("Setting up routes...")
	r := mux.NewRouter()
	r.HandleFunc("/ws/{roomUUID}/{userUUID}", enableCors(handleConnections))
	r.HandleFunc("/createRoom", enableCors(createRoom(database)))
	r.HandleFunc("/joinRoom", enableCors(joinRoom(database)))
	r.HandleFunc("/leaveRoom", enableCors(leaveRoom(database)))
	// r.HandleFunc("/vote", enableCors(vote(database)))
	r.HandleFunc("/showCards", enableCors(showCards(database)))
	r.HandleFunc("/autoShowCards", enableCors(autoShowCards(database)))
	r.HandleFunc("/resetVotes", enableCors(resetVotes(database)))
	r.HandleFunc("/changeName", enableCors(changeName(database)))

	log.Println("Starting server on " + os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), r))
}
