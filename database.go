package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

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
