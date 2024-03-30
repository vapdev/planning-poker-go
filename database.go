package main

import (
	"database/sql"
	"fmt"
	"log"
)

func createTables(database *sql.DB) {
	statements := []string{
		"CREATE TABLE IF NOT EXISTS rooms (id INTEGER PRIMARY KEY, slug TEXT)",
		"CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT, admin BOOLEAN)",
		"CREATE TABLE IF NOT EXISTS room_users (room_id INTEGER, user_id INTEGER, admin BOOLEAN, FOREIGN KEY(room_id) REFERENCES rooms(id), FOREIGN KEY(user_id) REFERENCES users(id))",
		"CREATE TABLE IF NOT EXISTS votes (room_id INTEGER, user_id INTEGER, vote INTEGER, FOREIGN KEY(room_id) REFERENCES rooms(id), FOREIGN KEY(user_id) REFERENCES users(id))",
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
