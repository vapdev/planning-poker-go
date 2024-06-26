package main

import (
	"log"

	"github.com/gorilla/websocket"
)

func handleLeaveRoom(game *Game, userID int) {
	var newAdmin *Player
	for i, player := range game.Players {
		if player.ID == userID {
			// Remove player from the game's players
			game.Players = append(game.Players[:i], game.Players[i+1:]...)

			// If the player is the admin, remember to assign a new one
			if player.Admin && len(game.Players) > 0 {
				newAdmin = game.Players[0]
			}

			break
		}
	}

	// If a new admin was chosen, assign the admin role to them
	if newAdmin != nil {
		newAdmin.Admin = true
	}
}

func handleVote(msg map[string]interface{}, game *Game, userID int) {
	log.Printf("Handling vote: %v", msg)
	vote, ok := msg["vote"].(float64)
	if !ok {
		log.Printf("vote is not a float64: %v", msg["vote"])
		return
	}

	voteInt := int(vote)

	// get db
	db := getDB()
	castVote(db, game.roomID, userID, voteInt)

	for _, player := range game.Players {
		if player.ID == userID {
			if player.Voted && player.Vote != nil && *player.Vote == voteInt {
				player.Voted = false
				player.Vote = nil
			} else {
				player.Voted = true
				player.Vote = &voteInt
			}
			break
		}
	}
}

func handleNewPlayer(msg map[string]interface{}, game *Game, userID int, userUUID string, ws *websocket.Conn) {
	name, ok := msg["name"].(string)
	if !ok || name == "" {
		name = ""
	}
	// Check if the user already exists in the game's players
	for _, player := range game.Players {
		if player.ID == userID {
			log.Printf("User %d already exists in the game", userID)
			return
		}
	}

	isAdmin := false
	if len(game.Players) == 0 {
		isAdmin = true
	}

	player := &Player{
		ID:    userID,
		UUID:  userUUID,
		Name:  name,
		Score: 0,
		Voted: false,
		Admin: isAdmin,
		ws:    ws,
	}
	game.Players = append(game.Players, player)
}

func handleNewAdmin(msg map[string]interface{}, game *Game, userID int, userUUID string, ws *websocket.Conn) {
	name, ok := msg["name"].(string)
	if !ok || name == "" {
		name = ""
	}

	player := &Player{
		ID:    userID,
		UUID:  userUUID,
		Name:  name,
		Score: 0,
		Voted: false,
		Admin: true,
		ws:    ws,
	}
	game.Players = append(game.Players, player)
}
