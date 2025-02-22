package main

import (
	"database/sql"
	"log"

	"github.com/gorilla/websocket"
)

func handleLeaveRoom(game *Game, userID int) {
	var newAdmin *Player
	for i, player := range game.Players {
		if player.ID == userID {
			//check if Player has connections
			if len(player.connections) > 1 {
				//remove the connection from the player
				for j, conn := range player.connections {
					if conn != nil {
						err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
						if err != nil {
							log.Printf("Error closing connection for player %d: %v", player.ID, err)
						}
						player.connections[j] = nil
					}
				}
			}
			//remove the player from the game
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

func handleVote(msg map[string]interface{}, game *Game, userID int, db *sql.DB) {
	vote, ok := msg["vote"].(string)
	if !ok {
		log.Println("Invalid vote format")
		return
	}

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

func handleIssueOrder(msg map[string]interface{}, game *Game, db *sql.DB) {
	issues, ok := msg["issues"].([]interface{})
	if !ok {
		log.Println("Invalid issue format")
		log.Println("issueData", issues)
		return
	}

	for i, issue := range issues {
		issueUUID, ok := issue.(string)
		if !ok {
			log.Println("Invalid issue format")
			log.Println("issue", issue)
			return
		}
		// get issue whole object  by uuid
		issue, err := getIssueByUUID(db, game.roomID, issueUUID)
		if err != nil {
			log.Println("Error getting issue by UUID")
			return
		}

		err = updateIssueOrder(db, game.roomID, issue["id"].(int), i)
		if err != nil {
			log.Println("Error updating issue order")
			return
		}
		game.issues[i] = Issue{
			ID:          issue["id"].(int),
			UUID:        issueUUID,
			Title:       issue["title"].(string),
			Description: issue["description"].(string),
			Link:        issue["link"].(string),
			Sequence:    i,
		}

		if err != nil {
			log.Println("Error updating issue order")
			return
		}
	}
}

func handleNewIssue(msg map[string]interface{}, game *Game, db *sql.DB) {
	issueData, ok := msg["issue"].(map[string]interface{})
	if !ok {
		log.Println("Invalid issue format")
		return
	}

	title, _ := issueData["title"].(string)
	description, _ := issueData["description"].(string)
	link, _ := issueData["link"].(string)

	uuid := generateUuid()
	id, err := createIssue(db, game.roomID, uuid, IssueRequest{
		RoomID:      game.roomID,
		Title:       title,
		Description: description,
		Link:        link,
		Sequence:    len(game.issues),
	})

	if err != nil {
		log.Printf("Error creating issue: %v", err)
		return
	}

	issue := Issue{
		ID:          int(id),
		UUID:        uuid,
		Title:       title,
		Description: description,
		Link:        link,
	}

	game.issues = append(game.issues, issue)
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

	emojiMessage := EmojiMessage{
		Emoji:        emoji,
		OriginUserID: userID,
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
			player.connections = append(player.connections, ws)
			return
		}
	}

	isAdmin := false
	if len(game.Players) == 0 {
		isAdmin = true
	}

	player := &Player{
		ID:          userID,
		UUID:        userUUID,
		Name:        name,
		Score:       0,
		Voted:       false,
		Admin:       isAdmin,
		connections: []*websocket.Conn{ws},
	}
	game.Players = append(game.Players, player)
}

func handleNewAdmin(msg map[string]interface{}, game *Game, userID int, userUUID string, ws *websocket.Conn) {
	name, ok := msg["name"].(string)
	if !ok || name == "" {
		name = ""
	}

	// Check if the user already exists in the game's players
	// If the user already exists, add the new connection to the player
	// If the user doesn't exist, create a new player and add it to the game
	for _, player := range game.Players {
		if player.ID == userID {
			log.Printf("User %d already exists in the game", userID)
			player.Admin = true
			//check if the connection already exists
			connectionExists := false
			for _, conn := range player.connections {
				if conn == ws {
					connectionExists = true
					break
				}
			}
			if !connectionExists {
				player.connections = append(player.connections, ws)
			}
			return
		}
	}

	player := &Player{
		ID:          userID,
		UUID:        userUUID,
		Name:        name,
		Score:       0,
		Voted:       false,
		Admin:       true,
		connections: []*websocket.Conn{ws},
	}

	game.Players = append(game.Players, player)
}
