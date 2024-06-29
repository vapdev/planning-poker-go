package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
)

func setupDatabase() *sql.DB {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	database, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	createTables(database)
	return database
}

func createTables(database *sql.DB) {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS rooms (
			id SERIAL PRIMARY KEY, 
			uuid UUID, 
			name TEXT, 
			showCards BOOLEAN DEFAULT FALSE, 
			autoShowCards BOOLEAN DEFAULT FALSE, 
			admin INTEGER
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY, 
			name TEXT, 
			uuid UUID, 
			guest BOOLEAN DEFAULT TRUE
		)`,
		`CREATE TABLE IF NOT EXISTS room_users (
			room_id INTEGER, 
			user_id INTEGER, 
			FOREIGN KEY(room_id) REFERENCES rooms(id), 
			FOREIGN KEY(user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS votes (
			room_id INTEGER, 
			user_id INTEGER, 
			vote INTEGER, 
			FOREIGN KEY(room_id) REFERENCES rooms(id), 
			FOREIGN KEY(user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS rounds (
			id SERIAL PRIMARY KEY, 
			room_id INTEGER, 
			start_time TIMESTAMP, 
			end_time TIMESTAMP, 
			finished BOOLEAN DEFAULT FALSE, 
			sequential INTEGER, 
			FOREIGN KEY(room_id) REFERENCES rooms(id)
		)`,
	}

	for _, stmt := range statements {
		statement, err := database.Prepare(stmt)
		if err != nil {
			log.Fatal(err)
		}
		_, err = statement.Exec()
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Tables created successfully.")
}
