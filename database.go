package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
)

func setupDatabase() *sql.DB {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file")
	}
	database, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	createTables(database)
	return database
}

func fetchGameFromDB(db *sql.DB, roomUUID string) (*Game, error) {
	query := `
		SELECT 
			r.id, r.uuid, r.name, r.showCards, r.autoShowCards, r.admin, r.lastActive
		FROM 
			rooms r
		WHERE 
			r.uuid = $1
	`

	var game Game
	var lastActive sql.NullTime

	err := db.QueryRow(query, roomUUID).Scan(
		&game.roomID,
		&game.roomUUID,
		&game.name,
		&game.showCards,
		&game.autoShowCards,
		&game.admin,
		&lastActive,
	)
	if err != nil {
		return nil, fmt.Errorf("error fetching game from DB: %v", err)
	}

	if lastActive.Valid {
		game.lastActive = lastActive.Time
	} else {
		game.lastActive = time.Now() // Or set to a default time
	}

	players, err := fetchPlayersFromDB(db, game.roomID)
	if err != nil {
		return nil, fmt.Errorf("error fetching players from DB: %v", err)
	}
	game.Players = players

	return &game, nil
}

func fetchPlayersFromDB(db *sql.DB, roomID int) ([]*Player, error) {
	query := `
		SELECT 
			u.id, u.uuid, u.name
		FROM 
			users u
		JOIN 
			room_users ru ON u.id = ru.user_id
		WHERE 
			ru.room_id = $1
	`

	rows, err := db.Query(query, roomID)
	if err != nil {
		return nil, fmt.Errorf("error fetching players from DB: %v", err)
	}
	defer rows.Close()

	var players []*Player
	for rows.Next() {
		var player Player
		err := rows.Scan(
			&player.ID,
			&player.UUID,
			&player.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning player from DB: %v", err)
		}
		players = append(players, &player)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating players rows: %v", err)
	}

	return players, nil
}

func createTables(database *sql.DB) {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS rooms (
			id SERIAL PRIMARY KEY, 
			uuid UUID, 
			name varchar(255), 
			showCards BOOLEAN DEFAULT FALSE, 
			autoShowCards BOOLEAN DEFAULT FALSE, 
			deck TEXT,
			admin INTEGER,
			lastActive TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY, 
			name varchar(255), 
			uuid UUID, 
			guest BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS room_users (
			room_id INTEGER, 
			user_id INTEGER, 
			FOREIGN KEY(room_id) REFERENCES rooms(id), 
			FOREIGN KEY(user_id) REFERENCES users(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS votes (
			room_id INTEGER, 
			user_id INTEGER, 
			vote varchar(5), 
			FOREIGN KEY(room_id) REFERENCES rooms(id), 
			FOREIGN KEY(user_id) REFERENCES users(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS rounds (
			id SERIAL PRIMARY KEY, 
			room_id INTEGER, 
			start_time TIMESTAMP, 
			end_time TIMESTAMP, 
			finished BOOLEAN DEFAULT FALSE, 
			sequential INTEGER, 
			FOREIGN KEY(room_id) REFERENCES rooms(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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