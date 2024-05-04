package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gocql/gocql"
	"github.com/golang-jwt/jwt/v5"
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

func FindUserByEmail(email string) (user.User, error) {
	var User user.User
	query := "SELECT user_id, username, email, password, bio, created_at FROM users WHERE email = ? LIMIT 1 ALLOW FILTERING"
	if err := db.Session.Query(query, email).Scan(&User.UserID, &User.Username, &User.Email, &User.Password, &User.Bio, &User.CreatedAt); err != nil {
		log.Println("Error while querying database:", err)
		return User, err
	}
	return User, nil
}

func FindUserByID(user_id string) (user.User, error) {
	var User user.User
	query := "SELECT user_id, username, email, password, bio, created_at FROM users WHERE user_id = ? LIMIT 1 ALLOW FILTERING"
	if err := db.Session.Query(query, user_id).Scan(&User.UserID, &User.Username, &User.Email, &User.Password, &User.Bio, &User.CreatedAt); err != nil {
		log.Println("Error while querying database:", err)
		return User, err
	}
	return User, nil
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
	existingUser, err = FindUserByEmail(newUser.Email)
	if err == nil {
		http.Error(w, "Email already exists", http.StatusConflict)
		log.Println("Email already exists in the database")
		return
	} else if err != gocql.ErrNotFound {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	newUser.CreatedAt = time.Now()
	newUser.UserID = gocql.MustRandomUUID().String()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error hashing password:", err)
		return
	}
	newUser.Password = string(hashedPassword)

	query := "INSERT INTO users (user_id, username, email, password, created_at) VALUES (?, ?, ?, ?, ?)"

	if err := db.Session.Query(query, newUser.UserID, newUser.Username, newUser.Email, newUser.Password, newUser.CreatedAt).Exec(); err != nil {
		log.Println("Error while inserting into database:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Println("received request to create a User")

	// generate access token and refresh token
	accessToken, refreshToken, err := CreateToken(existingUser.UserID)
	if err != nil {
		log.Println("Error generating token!")
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
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
	existingUser, err = FindUserByEmail(loginDetails.Email)
	if err != nil {
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
	accessToken, refreshToken, err := CreateToken(existingUser.UserID)
	if err != nil {
		log.Println("Error generating token!")
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
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
	type RefreshTokenReq struct {
		RefreshToken string `json:"refreshToken"`
	}
	var ReqBody RefreshTokenReq
	err := json.NewDecoder(r.Body).Decode(&ReqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	refreshToken := ReqBody.RefreshToken
	if refreshToken == "" {
		http.Error(w, "Missing refresh token.", http.StatusBadRequest)
		return
	}
	token, err := verifyToken(refreshToken)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error while varifying token", http.StatusUnauthorized)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if userID, ok := claims["user_id"].(string); ok {
			user, err := FindUserByID(userID)
			if err != nil {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			accessToken, refreshToken, err := CreateToken(user.UserID)
			if err != nil {
				log.Println("Error generating token!")
				http.Error(w, "Error generating token", http.StatusInternalServerError)
				return
			}
			response := map[string]interface{}{"message": "Token refreshed!", "accessToken": accessToken, "refreshToken": refreshToken}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				log.Println("Error encoding JSON response:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "Invalid token claims", http.StatusBadRequest)
		return
	}
}
