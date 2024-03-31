package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type User struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Bio       string    `json:"user_bio"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := "INSERT INTO users (user_id, username, email, password, created_at, created_channels, bio) VALUES (?, ?, ?, ?, ?, ?, ?)"

	if err := Session.Query(query, newUser.UserID, newUser.Username, newUser.Email, newUser.Password, newUser.CreatedAt, newUser.Bio).Exec(); err != nil {
		log.Println("Error while inserting")
		log.Println(err)
	}

	w.WriteHeader(http.StatusCreated)
	log.Println("received request to create a User")
	n, err := w.Write([]byte("User created!"))
	_ = n
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	initDb()
	router := http.NewServeMux()

	router.HandleFunc("POST /api/user", handleCreate)
	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	log.Printf("Starting server on port %s", server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
