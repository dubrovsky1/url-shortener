package main

import (
	"github.com/dubrovsky1/url-shortener/internal/handlers"
	"github.com/dubrovsky1/url-shortener/internal/storage"

	"log"
	"net/http"
)

func main() {
	h := handlers.Handler{*storage.New()}
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, h.MainHandler)

	log.Println("Server is listening localhost:8080")
	log.Fatal(http.ListenAndServe(`localhost:8080`, mux))
}
