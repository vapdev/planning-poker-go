package main

import "github.com/gorilla/websocket"

type Player struct {
	ID    int    `json:"id"`
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Score int    `json:"score"`
	Voted bool   `json:"voted"`
	Vote  int    `json:"vote"`
	Admin bool   `json:"admin"`
	ws    *websocket.Conn
}

type Game struct {
	Players       []*Player
	admin         int
	showCards     bool
	autoShowCards bool
	roomID        int
}
