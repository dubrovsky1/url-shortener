package main

import (
	"github.com/dubrovsky1/url-shortener/internal/config"
	"github.com/dubrovsky1/url-shortener/internal/handlers/api/shorten"
	"github.com/dubrovsky1/url-shortener/internal/handlers/api/user"
	"github.com/dubrovsky1/url-shortener/internal/handlers/geturl"
	"github.com/dubrovsky1/url-shortener/internal/handlers/ping"
	"github.com/dubrovsky1/url-shortener/internal/handlers/saveurl"
	"github.com/dubrovsky1/url-shortener/internal/middleware/auth"
	"github.com/dubrovsky1/url-shortener/internal/middleware/gzip"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/service"
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
	defer storage.Close()

	//создаем объект стоя бизнес-логики, который взаимодействует с базой
	serv := service.New(storage)

	r := chi.NewRouter()
	r.Post("/", auth.Auth(logger.WithLogging(gzip.GzipMiddleware(saveurl.SaveURL(serv, flags.ResultShortURL)))))
	r.Post("/api/shorten", auth.Auth(logger.WithLogging(gzip.GzipMiddleware(shorten.Shorten(serv, flags.ResultShortURL)))))
	r.Post("/api/shorten/batch", auth.Auth(logger.WithLogging(gzip.GzipMiddleware(shorten.Batch(serv)))))
	r.Get("/{id}", logger.WithLogging(gzip.GzipMiddleware(geturl.GetURL(serv))))
	r.Get("/ping", logger.WithLogging(gzip.GzipMiddleware(ping.Ping(flags.ConnectionString))))
	r.Get("/api/user/urls", auth.Auth(logger.WithLogging(gzip.GzipMiddleware(user.ListByUserId(serv)))))

	logger.Sugar.Infow("Server is listening", "host", flags.Host)
	log.Fatal(http.ListenAndServe(flags.Host, r))
}
