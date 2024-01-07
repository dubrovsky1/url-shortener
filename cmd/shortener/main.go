package main

import (
	"github.com/dubrovsky1/url-shortener/internal/config"
	"github.com/dubrovsky1/url-shortener/internal/handlers"
	"github.com/dubrovsky1/url-shortener/internal/middleware/gzip"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
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

	logger.Initialize()

	r := chi.NewRouter()
	r.Post("/", logger.WithLogging(gzip.GzipMiddleware(h.SaveURL)))
	r.Post("/api/shorten", logger.WithLogging(gzip.GzipMiddleware(h.Shorten)))
	r.Get("/{id}", logger.WithLogging(gzip.GzipMiddleware(h.GetURL)))

	logger.Sugar.Infow("Flags:", "-a", flags.Host, "-b", flags.ResultShortURL)
	logger.Sugar.Infow("Server is listening", "host", flags.Host)

	log.Fatal(http.ListenAndServe(flags.Host, r))
}
