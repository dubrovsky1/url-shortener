package main

import (
	"github.com/dubrovsky1/url-shortener/internal/config"
	"github.com/dubrovsky1/url-shortener/internal/handlers"
	"github.com/dubrovsky1/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	//парсим переменные окружения и флаги из конфигуратора
	flags := config.ParseFlags()
	//хендлер с доступом к хранилищу
	h := handlers.New(flags.ResultShortURL, storage.New())

	r := chi.NewRouter()
	r.Post("/", h.SaveURL)
	r.Get("/{id}", h.GetURL)

	log.Printf("Flags: -a %s, -b %s\n", flags.Host, flags.ResultShortURL)
	log.Printf("Server is listening %s\n", flags.Host)
	log.Fatal(http.ListenAndServe(flags.Host, r))
}
