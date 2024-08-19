package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	db_helpers "github.com/MrSantamaria/go3tier/db_helpers"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func main() {
	// Database connection
	db, err := sql.Open("mysql", os.Getenv("DB_USER")+":"+os.Getenv("DB_PASSWORD")+"@tcp("+os.Getenv("DB_HOST")+")/email_queue")
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Printf("Connected to database %s", os.Getenv("DB_HOST"))
	defer db.Close()

	// Initialize db in db_helpers package
	db_helpers.InitDB(db)

	// Create table if not exists
	db_helpers.CreateTable()

	// Set up router
	r := mux.NewRouter()
	r.HandleFunc("/", serveForm).Methods("GET")
	r.HandleFunc("/queue", db_helpers.GetQueue).Methods("GET")
	r.HandleFunc("/queue", addToQueueHTMX).Methods("POST")
	r.HandleFunc("/queue-view", serveQueueView).Methods("GET")
	r.HandleFunc("/queue/{id}", db_helpers.RemoveFromQueue).Methods("DELETE")

	// Start server
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// serveForm handles serving the HTMX-enabled HTML form
func serveForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Email Queue</title>
			<script src="https://unpkg.com/htmx.org@1.9.2"></script>
		</head>
		<body>
			<h1>Email Queue</h1>
			<form hx-post="/queue" hx-target="#result" hx-swap="innerHTML">
				<label for="name">Name:</label>
				<input type="text" id="name" name="name" required><br><br>
				<label for="email">Email:</label>
				<input type="email" id="email" name="email" required><br><br>
				<button type="submit">Submit</button>
			</form>
			<div id="result"></div>
			<h2>Current Queue</h2>
			<div id="queue" hx-get="/queue-view" hx-trigger="load,htmx:afterRequest" hx-swap="innerHTML"></div>
		</body>
		</html>
	`)
}

// addToQueueHTMX handles the HTMX-enabled form submission
func addToQueueHTMX(w http.ResponseWriter, r *http.Request) {
	var eq db_helpers.EmailQueue
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	eq.Name = r.FormValue("name")
	eq.Email = r.FormValue("email")

	_, err := db_helpers.DB.Exec("INSERT INTO email_queue (name, email) VALUES (?, ?)", eq.Name, eq.Email)
	if err != nil {
		http.Error(w, "Failed to insert into database", http.StatusInternalServerError)
		return
	}

	// Return a partial HTML snippet that HTMX will inject into the page
	fmt.Fprintf(w, "<p>Thank you, %s! Your email (%s) has been added to the queue.</p>", eq.Name, eq.Email)
}

// serveQueueView returns the current queue as an HTML snippet
func serveQueueView(w http.ResponseWriter, r *http.Request) {
	rows, err := db_helpers.DB.Query("SELECT id, name, email FROM email_queue")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	fmt.Fprintf(w, "<ul>")
	for rows.Next() {
		var id int
		var name, email string
		if err := rows.Scan(&id, &name, &email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "<li>%s (%s)</li>", name, email)
	}
	fmt.Fprintf(w, "</ul>")
}
