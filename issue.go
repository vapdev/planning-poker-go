package main

import (
	"database/sql"
)

type IssueRequest struct {
	RoomID      int    `json:"room_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Sequence    int    `json:"sequence"`
}

func createIssue(database *sql.DB, roomId int, uuid string, issue IssueRequest) (int64, error) {
	var id int64
	err := database.QueryRow("INSERT INTO issues (room_id, uuid, title, description, link, sequence) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		roomId, uuid, issue.Title, issue.Description, issue.Link, issue.Sequence).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func updateIssueOrder(database *sql.DB, roomId, issueId, sequence int) error {
	_, err := database.Exec("UPDATE issues SET sequence = $1 WHERE room_id = $2 AND id = $3", sequence, roomId, issueId)
	if err != nil {
		return err
	}

	return nil
}

func getIssueByUUID(db *sql.DB, roomID int, issueUUID string) (map[string]interface{}, error) {
	var id int
	var title, description, link string

	err := db.QueryRow("SELECT id, title, description, link FROM issues WHERE room_id = $1 AND uuid = $2", roomID, issueUUID).
		Scan(&id, &title, &description, &link)
	if err != nil {
		return nil, err
	}

	// Criar o mapa manualmente
	issue := map[string]interface{}{
		"id":          id,
		"title":       title,
		"description": description,
		"link":        link,
	}

	return issue, nil
}
