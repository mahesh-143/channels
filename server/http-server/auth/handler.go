package auth

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mahesh-143/channels/db"
	"github.com/mahesh-143/channels/user"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct{}

type LoginDetails struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
	accessToken, err := createToken(existingUser.Username)
	if err != nil {
		log.Println("Error generating token!")
	}

	w.WriteHeader(http.StatusOK)
	log.Println(w, "Login Successfull")

	userDetails := existingUser
	userDetails.Password = ""

	response := map[string]interface{}{"message": "Login successfull!", "accessToken": accessToken, "user": userDetails}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error encoding JSON response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// refresh token
}
