package main

import (
	"github.com/dubrovsky1/url-shortener/internal/config"
	"github.com/dubrovsky1/url-shortener/internal/handlers/batch"
	"github.com/dubrovsky1/url-shortener/internal/handlers/geturl"
	"github.com/dubrovsky1/url-shortener/internal/handlers/ping"
	"github.com/dubrovsky1/url-shortener/internal/handlers/saveurl"
	"github.com/dubrovsky1/url-shortener/internal/handlers/shorten"
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

	logger.Initialize()
	logger.Sugar.Infow("Flags:", "-a", flags.Host, "-b", flags.ResultShortURL, "-f", flags.FileStoragePath, "-d", flags.ConnectionString)

	storage, err := storage.GetStorage(flags)
	if err != nil {
		log.Fatal("Get storage error. ", err)
	}

	r := chi.NewRouter()
	r.Post("/", logger.WithLogging(gzip.GzipMiddleware(saveurl.SaveURL(storage, flags.ResultShortURL))))
	r.Post("/api/shorten", logger.WithLogging(gzip.GzipMiddleware(shorten.Shorten(storage, flags.ResultShortURL))))
	r.Post("/api/shorten/batch", logger.WithLogging(gzip.GzipMiddleware(batch.Batch(storage))))
	r.Get("/{id}", logger.WithLogging(gzip.GzipMiddleware(geturl.GetURL(storage))))
	r.Get("/ping", logger.WithLogging(gzip.GzipMiddleware(ping.Ping(flags.ConnectionString))))

	logger.Sugar.Infow("Server is listening", "host", flags.Host)
	log.Fatal(http.ListenAndServe(flags.Host, r))
}
