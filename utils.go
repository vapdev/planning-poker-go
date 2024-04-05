package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
)

type Palavra struct {
	Nome   string
	Genero string
}

var adjetivos = []Palavra{
	{"assustador", "masculino"},
	{"assustadora", "feminino"},
	{"engraçado", "masculino"},
	{"engraçada", "feminino"},
	{"doido", "masculino"},
	{"doida", "feminino"},
	{"observável", "masculino"},
	{"observável", "feminino"},
	{"esquisito", "masculino"},
	{"esquisita", "feminino"},
	{"colorido", "masculino"},
	{"colorida", "feminino"},
	{"gigante", "masculino"},
	{"gigante", "feminino"},
	{"pequeno", "masculino"},
	{"pequena", "feminino"},
	{"brilhante", "masculino"},
	{"brilhante", "feminino"},
	{"furioso", "masculino"},
	{"furiosa", "feminino"},
	{"tranquilo", "masculino"},
	{"tranquila", "feminino"},
	{"fantástico", "masculino"},
	{"fantástica", "feminino"},
	{"invisível", "masculino"},
	{"invisível", "feminino"},
	{"misterioso", "masculino"},
	{"misteriosa", "feminino"},
	{"fofo", "masculino"},
	{"fofa", "feminino"},
	{"maluco", "masculino"},
	{"maluca", "feminino"},
	{"cheiroso", "masculino"},
	{"cheirosa", "feminino"},
	{"inteligente", "feminino"},
	{"luminoso", "masculino"},
	{"luminosa", "feminino"},
	{"sedoso", "masculino"},
	{"sedosa", "feminino"},
	{"encantado", "masculino"},
	{"encantada", "feminino"},
	{"saboroso", "masculino"},
	{"saborosa", "feminino"},
	{"hilário", "masculino"},
	{"hilária", "feminino"},
	{"mal-humorado", "masculino"},
	{"mal-humorada", "feminino"},
	{"radiante", "masculino"},
	{"radiante", "feminino"},
	{"apaixonado", "masculino"},
	{"apaixonada", "feminino"},
	{"confuso", "masculino"},
	{"confusa", "feminino"},
	{"engenhoso", "masculino"},
	{"engenhosa", "feminino"},
	{"estranho", "masculino"},
	{"estranha", "feminino"},
	{"bizarro", "masculino"},
	{"bizarra", "feminino"},
}

var substantivos = []Palavra{
	{"morango", "masculino"},
	{"sabonete", "masculino"},
	{"cadeira", "feminino"},
	{"banana", "feminino"},
	{"cachorro", "masculino"},
	{"gato", "masculino"},
	{"papagaio", "masculino"},
	{"computador", "masculino"},
	{"celular", "masculino"},
	{"carro", "masculino"},
	{"bolo", "masculino"},
	{"chapéu", "masculino"},
	{"livro", "masculino"},
	{"brinquedo", "masculino"},
	{"caneca", "feminino"},
	{"jardim", "masculino"},
	{"copo", "masculino"},
	{"relógio", "masculino"},
	{"avião", "masculino"},
	{"barco", "masculino"},
	{"colher", "feminino"},
	{"espelho", "masculino"},
	{"sapato", "masculino"},
	{"porta", "feminino"},
	{"piano", "masculino"},
	{"telefone", "masculino"},
	{"janela", "feminino"},
	{"máquina", "feminino"},
	{"teclado", "masculino"},
	{"almofada", "feminino"},
}

func generateNomeSala(db *sql.DB) (string, error) {
	// Fetch all used names from the database.
	rows, err := db.Query("SELECT uuid FROM rooms")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// Store the used names in a map for quick lookup.
	usedNames := make(map[string]bool)
	for rows.Next() {
		var uuid string
		if err := rows.Scan(&uuid); err != nil {
			return "", err
		}
		usedNames[uuid] = true
	}

	// Generate a random name that is not used.
	var nomeSala string
	for {
		substantivo := substantivos[rand.Intn(len(substantivos))]
		adjetivo := adjetivos[rand.Intn(len(adjetivos))]

		// Check if the substantivo and adjetivo genders match
		if substantivo.Genero == "masculino" && adjetivo.Genero == "feminino" {
			continue // Skip this iteration, as genders don't match
		}
		if substantivo.Genero == "feminino" && adjetivo.Genero == "masculino" {
			continue // Skip this iteration, as genders don't match
		}

		// If genders match or both are neutral, proceed
		nomeSala = fmt.Sprintf("%s %s", substantivo.Nome, adjetivo.Nome)

		if !usedNames[nomeSala] {
			break
		}
	}

	return nomeSala, nil
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
