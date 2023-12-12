package main

import (
	"github.com/dubrovsky1/url-shortener/internal/handlers"
	"github.com/dubrovsky1/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"

	"log"
	"net/http"
)

func main() {
	//mux.HandleFunc(`/`, h.MainHandler)
	//mux := http.NewServeMux()
	//mux.HandleFunc(`/`, h.MainHandler)

	h := handlers.Handler{Urls: *storage.New()}
	r := chi.NewRouter()

	r.Post("/", h.SaveURL)
	r.Get("/{id}", h.GetURL)

	log.Println("Server is listening localhost:8080")
	log.Fatal(http.ListenAndServe(`localhost:8080`, r))
}
