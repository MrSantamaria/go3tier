package db_helpers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var DB *sql.DB

func InitDB(db *sql.DB) {
	DB = db
}

func CreateTable() {
	query := `
	CREATE TABLE IF NOT EXISTS email_queue (
		id INT AUTO_INCREMENT,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) NOT NULL,
		PRIMARY KEY (id)
	);`
	_, err := DB.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func GetQueue(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query("SELECT id, name, email FROM email_queue")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var queue []EmailQueue
	for rows.Next() {
		var eq EmailQueue
		if err := rows.Scan(&eq.ID, &eq.Name, &eq.Email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		queue = append(queue, eq)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(queue)
}

func AddToQueue(w http.ResponseWriter, r *http.Request) {
	var eq EmailQueue
	if err := json.NewDecoder(r.Body).Decode(&eq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := DB.Exec("INSERT INTO email_queue (name, email) VALUES (?, ?)", eq.Name, eq.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func RemoveFromQueue(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	_, err := DB.Exec("DELETE FROM email_queue WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type EmailQueue struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
