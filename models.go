package main

import (
	"time"

	"github.com/gorilla/websocket"
)

type Player struct {
	ID          int     `json:"id"`
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Score       int     `json:"score"`
	Voted       bool    `json:"voted"`
	Vote        *string `json:"vote"`
	Admin       bool    `json:"admin"`
	connections []*websocket.Conn
}

type CardOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type Issue struct {
	ID          int    `json:"id"`
	UUID        string `json:"uuid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Sequence    int    `json:"sequence"`
}

type Game struct {
	Players       []*Player
	name          string
	admin         int
	showCards     bool
	autoShowCards bool
	roomID        int
	roomUUID      string
	lastActive    time.Time
	Emojis        []EmojiMessage
	deck          []CardOption
	issues        []Issue
}

type EmojiMessage struct {
	Emoji        string
	OriginUserID int
	TargetUserID int
}
