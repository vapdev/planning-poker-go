package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	defer func() {
		log.Println("Cleanup routine stopping...")
	}()
	
	for {
		select {
		case <-cleanupTicker.C:
			now := time.Now()
			for roomUUID, game := range games {
				if now.Sub(game.lastActive) > roomExpiry {
					ammountOfPlayers := len(game.Players)
					if ammountOfPlayers == 0 {
						log.Printf("Cleaning up room %s", roomUUID)
						delete(games, roomUUID)
					}
				}
			}
		}
	}
}

func main() {
	// Setup database connection
	database := setupDatabase()
	
	// Ensure database connection is closed when the application exits
	defer func() {
		log.Println("Closing database connection...")
		if err := database.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		} else {
			log.Println("Database connection closed successfully")
		}
	}()
	
	log.Println("Setting up routes...")
	r := mux.NewRouter()

	r.HandleFunc("/ws/{roomUUID}/{userUUID}", enableCors(handleConnections(database)))
	r.HandleFunc("/createRoom", enableCors(createRoom(database)))
	r.HandleFunc("/joinRoom", enableCors(joinRoom(database)))
	r.HandleFunc("/leaveRoom", enableCors(leaveRoom(database)))
	r.HandleFunc("/showCards", enableCors(showCards(database)))
	r.HandleFunc("/autoShowCards", enableCors(autoShowCards(database)))
	r.HandleFunc("/resetVotes", enableCors(resetVotes(database)))
	r.HandleFunc("/changeName", enableCors(changeName(database)))
	r.HandleFunc("/changeRoomName", enableCors(changeRoomName(database)))
	r.HandleFunc("/kickPlayer", enableCors(kickPlayer(database)))

	// Start cleanup routine in a goroutine
	cleanupDone := make(chan bool)
	go func() {
		defer func() {
			cleanupDone <- true
		}()
		cleanupExpiredRooms()
	}()

	// Create a server with timeout configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}
	
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the server in a goroutine
	go func() {
		log.Println("Starting server on port " + port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-stop
	
	// Stop the cleanup ticker
	cleanupTicker.Stop()
	
	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Attempt graceful shutdown
	log.Println("Shutting down server...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}
	
	// Wait for cleanup to finish
	select {
	case <-cleanupDone:
		log.Println("Cleanup routine finished")
	case <-time.After(5 * time.Second):
		log.Println("Cleanup routine timeout, forcing shutdown")
	}
	
	log.Println("Server gracefully stopped")
}
