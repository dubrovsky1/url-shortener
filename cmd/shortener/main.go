package main

import (
	"github.com/dubrovsky1/url-shortener/internal/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc(`/{id}`, handlers.GetHandler)
	router.HandleFunc(`/`, handlers.PostHandler)

	log.Println("Server is listening localhost:8080")
	log.Fatal(http.ListenAndServe(`localhost:8080`, router))
}
