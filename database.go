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
	database, err := sql.Open("sqlite3", "planningpoker.db")
	if err != nil {
		log.Fatal(err)
	}

	statements := []string{
		"CREATE TABLE IF NOT EXISTS rooms (id INTEGER PRIMARY KEY, uuid TEXT, name TEXT, showCards BOOLEAN DEFAULT 0, autoShowCards BOOLEAN DEFAULT 0, admin INTEGER, FOREIGN KEY(admin) REFERENCES users(id))",
		"CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT, uuid TEXT, guest BOOLEAN DEFAULT 1)",
		"CREATE TABLE IF NOT EXISTS room_users (room_id INTEGER, user_id INTEGER, FOREIGN KEY(room_id) REFERENCES rooms(id), FOREIGN KEY(user_id) REFERENCES users(id))",
		"CREATE TABLE IF NOT EXISTS votes (room_id INTEGER, user_id INTEGER, vote INTEGER, FOREIGN KEY(room_id) REFERENCES rooms(id), FOREIGN KEY(user_id) REFERENCES users(id))",
		"CREATE TABLE IF NOT EXISTS rounds (id INTEGER PRIMARY KEY, room_id INTEGER, start_time TEXT, end_time TEXT, finished BOOLEAN DEFAULT 0, sequential INTEGER, FOREIGN KEY(room_id) REFERENCES rooms(id))",
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
