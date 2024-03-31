package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Message struct {
	Text string `json:"message"`
}

func handleHello(w http.ResponseWriter, r *http.Request) {
	message := Message{Text: "Hello, world!"}
	err := json.NewEncoder(w).Encode(message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	router := http.NewServeMux()
	router.HandleFunc("GET /hello", handleHello)
	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	log.Printf("Starting server on port %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
