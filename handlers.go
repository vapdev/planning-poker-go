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
    vote, ok := msg["vote"].(string)
    if !ok {
        log.Println("Invalid vote format")
        return
    }

    // get db
    db := getDB()
    castVote(db, game.roomID, userID, vote)

    for _, player := range game.Players {
        if player.ID == userID {
            if player.Voted && player.Vote != nil && *player.Vote == vote {
                player.Voted = false
                player.Vote = nil
            } else {
                player.Voted = true
                player.Vote = &vote
            }
            break
        }
    }
}

func handleEmoji(msg map[string]interface{}, game *Game, userID int) {
	emoji, ok := msg["emoji"].(string)
	if !ok {
		log.Printf("emoji is not a string: %v", msg["emoji"])
		return
	}

	targetUserIdFloat, ok := msg["targetUserId"].(float64)
	if !ok {
		log.Printf("targetUserId is not a float64: %v", msg["targetUserId"])
		return
	}
	targetUserId := int(targetUserIdFloat)

	originUserIdFloat, ok := msg["originUserId"].(float64)
	if !ok {
		log.Printf("originUserId is not a float64: %v", msg["originUserId"])
		return
	}
	originUserId := int(originUserIdFloat)

	emojiMessage := EmojiMessage{
		Emoji:        emoji,
		OriginUserID: originUserId,
		TargetUserID: targetUserId,
	}

	// Send the game state with the emoji message
	sendGameState(game, []EmojiMessage{emojiMessage})
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
