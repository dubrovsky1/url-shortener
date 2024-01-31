package main

import (
	"github.com/dubrovsky1/url-shortener/internal/config"
	"github.com/dubrovsky1/url-shortener/internal/handlers/geturl"
	"github.com/dubrovsky1/url-shortener/internal/handlers/ping"
	"github.com/dubrovsky1/url-shortener/internal/handlers/saveurl"
	"github.com/dubrovsky1/url-shortener/internal/handlers/shorten"
	"github.com/dubrovsky1/url-shortener/internal/middleware/gzip"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/storage"
	"github.com/dubrovsky1/url-shortener/internal/storage/file"
	"github.com/dubrovsky1/url-shortener/internal/storage/memory"
	"github.com/dubrovsky1/url-shortener/internal/storage/postgresql"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	//парсим переменные окружения и флаги из конфигуратора
	flags := config.ParseFlags()

	logger.Initialize()
	logger.Sugar.Infow("Flags:", "-a", flags.Host, "-b", flags.ResultShortURL, "-f", flags.FileStoragePath, "-d", flags.ConnectionString)

	storage := getStorage(flags)

	r := chi.NewRouter()
	r.Post("/", logger.WithLogging(gzip.GzipMiddleware(saveurl.SaveURL(storage, flags.ResultShortURL))))
	r.Post("/api/shorten", logger.WithLogging(gzip.GzipMiddleware(shorten.Shorten(storage, flags.ResultShortURL))))
	r.Get("/{id}", logger.WithLogging(gzip.GzipMiddleware(geturl.GetURL(storage))))
	r.Get("/ping", logger.WithLogging(gzip.GzipMiddleware(ping.Ping(flags.ConnectionString))))

	logger.Sugar.Infow("Server is listening", "host", flags.Host)
	log.Fatal(http.ListenAndServe(flags.Host, r))
}

func getStorage(flags config.Config) storage.Repository {
	var db storage.Repository
	var err error

	if flags.ConnectionString != "" {
		db, err = postgresql.New(flags.ConnectionString)
		if err != nil {
			log.Fatal("Postgresql storage init error", err)
		}
	} else if flags.FileStoragePath != "" {
		db, err = file.New(flags.FileStoragePath)
		if err != nil {
			log.Fatal("File storage init error", err)
		}
	} else {
		db = memory.New()
	}
	return db
}
