package main

import (
	"log"
	"net/http"

	"github.com/mahesh-143/channels/auth"
	"github.com/mahesh-143/channels/db"
)

func main() {
	db.InitDb()
	auth_handler := &auth.Handler{}
	router := http.NewServeMux()

	router.HandleFunc("POST /api/user", auth_handler.Register)
	router.HandleFunc("POST /api/auth/login", auth_handler.Login)
	router.HandleFunc("POST /api/auth/refreshtoken", auth_handler.RefreshToken)

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
