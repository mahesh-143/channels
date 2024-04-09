package main

import (
	"log"
	"net/http"

	"github.com/mahesh-143/channels/auth"
	"github.com/mahesh-143/channels/db"
	"github.com/mahesh-143/channels/user"
)

func main() {
	db.InitDb()
	user_handler := &user.Handler{}
	auth_handler := &auth.Handler{}
	router := http.NewServeMux()

	router.HandleFunc("POST /api/user", user_handler.Create)
	router.HandleFunc("POST /api/auth/login", auth_handler.Login)

	server := http.Server{
		Addr: ":8080",

		Handler: router,
	}
	log.Printf("Starting server on port %s", server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
