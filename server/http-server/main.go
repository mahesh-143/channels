package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gocql/gocql"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserID    gocql.UUID `json:"user_id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Password  string     `json:"password"`
	CreatedAt time.Time  `json:"created_at"`
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Println("Error decoding request body: ", err)
		return
	}
	defer r.Body.Close()

	var existingUser User
	query := "SELECT user_id FROM users WHERE email = ? LIMIT 1 ALLOW FILTERING"
	if err := Session.Query(query, newUser.Email).Scan(&existingUser.UserID); err == nil {
		http.Error(w, "Email already exists", http.StatusConflict)
		log.Println("Email already exists in the database")
		return
	} else if err != gocql.ErrNotFound {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error querying database:", err)
		return
	}

	newUser.CreatedAt = time.Now()
	newUser.UserID = gocql.TimeUUID()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error hashing password:", err)
		return
	}
	newUser.Password = string(hashedPassword)

	query = "INSERT INTO users (user_id, username, email, password, created_at) VALUES (?, ?, ?, ?, ?)"

	if err := Session.Query(query, newUser.UserID, newUser.Username, newUser.Email, newUser.Password, newUser.CreatedAt).Exec(); err != nil {
		log.Println("Error while inserting into database:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Println("received request to create a User")

	response := map[string]string{"message": "User created!"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error encoding JSON response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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
