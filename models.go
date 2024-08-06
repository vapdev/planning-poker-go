package main

import "github.com/gorilla/websocket"
import "time"

type Player struct {
	ID    int    `json:"id"`
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Score int    `json:"score"`
	Voted bool   `json:"voted"`
	Vote  *string   `json:"vote"`
	Admin bool   `json:"admin"`
	ws    *websocket.Conn
}

type CardOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type Game struct {
	Players       []*Player
	name          string
	admin         int
	showCards     bool
	autoShowCards bool
	roomID        int
	roomUUID      string
	lastActive	  time.Time
	Emojis      []EmojiMessage
	deck        []CardOption
}

type EmojiMessage struct {
	Emoji        string
	OriginUserID int
	TargetUserID int
}
