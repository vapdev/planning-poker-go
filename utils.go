package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type Palavra struct {
	Nome   string
	Genero string
}

func generateRoomUUID(db *sql.DB) (string, error) {
	var uuidStr string
	uuidStr = uuid.New().String()
	return uuidStr, nil
}

func handleError(w http.ResponseWriter, err error) bool {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	return false
}

func sendResponse(w http.ResponseWriter, data map[string]interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func getUserIDFromUUID(db *sql.DB, uuid string) (int64, error) {
	var id int64
	err := db.QueryRow("SELECT id FROM users WHERE uuid = $1", uuid).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func getRoomIDFromUUID(db *sql.DB, uuid string) (int, error) {
	var id int
	err := db.QueryRow("SELECT id FROM rooms WHERE uuid = $1", uuid).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
