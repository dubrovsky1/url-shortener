package main

import (
	"github.com/dubrovsky1/url-shortener/internal/handlers"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handlers.MainHandler)

	log.Println("Server is listening localhost:8080")
	log.Fatal(http.ListenAndServe(`localhost:8080`, mux))
}
