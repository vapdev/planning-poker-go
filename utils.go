package main

import (
	"database/sql"
	"encoding/json"
	"math/rand"
	"net/http"
)

func generateRandomString(n int) string {
	var letters = []rune("1234567890")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
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
	err := db.QueryRow("SELECT id FROM users WHERE uuid = ?", uuid).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func getRoomIDFromUUID(db *sql.DB, uuid string) (int, error) {
	var id int
	err := db.QueryRow("SELECT id FROM rooms WHERE uuid = ?", uuid).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
