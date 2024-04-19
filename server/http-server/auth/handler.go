package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gocql/gocql"
	"github.com/mahesh-143/channels/db"
	"github.com/mahesh-143/channels/user"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct{}

type LoginDetails struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// return user without password
func UserDetails(u user.User) user.User {
	userDetails := u
	userDetails.Password = ""
	return userDetails
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var newUser user.User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Println("Error decoding request body: ", err)
		return
	}
	defer r.Body.Close()

	var existingUser user.User
	query := "SELECT user_id FROM users WHERE email = ? LIMIT 1 ALLOW FILTERING"
	if err := db.Session.Query(query, newUser.Email).Scan(&existingUser.UserID); err == nil {
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

	if err := db.Session.Query(query, newUser.UserID, newUser.Username, newUser.Email, newUser.Password, newUser.CreatedAt).Exec(); err != nil {
		log.Println("Error while inserting into database:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Println("received request to create a User")

	// generate access token and refresh token
	accessToken, refreshToken, err := CreateToken(existingUser.Username)
	if err != nil {
		log.Println("Error generating token!")
	}

	response := map[string]interface{}{"message": "User created!", "accessToken": accessToken, "refreshToken": refreshToken, "user": UserDetails(newUser)}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error encoding JSON response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var loginDetails LoginDetails
	err := json.NewDecoder(r.Body).Decode(&loginDetails)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Println("Error decoding request body: ", err)
		return
	}
	defer r.Body.Close()

	// find user by email
	var existingUser user.User
	query := "SELECT user_id, username, email, password, bio, created_at FROM users WHERE email = ? LIMIT 1 ALLOW FILTERING"
	if err := db.Session.Query(query, loginDetails.Email).Scan(&existingUser.UserID, &existingUser.Username, &existingUser.Email, &existingUser.Password, &existingUser.Bio, &existingUser.CreatedAt); err != nil {
		log.Println("Error while querying database:", err)
		http.Error(w, "Invalid Email or Password", http.StatusUnauthorized)
		return
	}

	// validate password
	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(loginDetails.Password)); err != nil {
		log.Println("Error while comparing hashed password: ", err)
		http.Error(w, "Invalid Email or Password", http.StatusUnauthorized)
		return
	}

	// generate access token and refresh token
	accessToken, refreshToken, err := CreateToken(existingUser.Username)
	if err != nil {
		log.Println("Error generating token!")
	}

	w.WriteHeader(http.StatusOK)
	log.Println(w, "Login Successfull")

	response := map[string]interface{}{"message": "Login successfull!", "accessToken": accessToken, "refreshToken": refreshToken, "user": UserDetails(existingUser)}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error encoding JSON response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// refresh token
}
