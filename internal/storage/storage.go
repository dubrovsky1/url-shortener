package storage

import (
	"context"
	"github.com/dubrovsky1/url-shortener/internal/config"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/dubrovsky1/url-shortener/internal/storage/file"
	"github.com/dubrovsky1/url-shortener/internal/storage/memory"
	"github.com/dubrovsky1/url-shortener/internal/storage/postgresql"
	"github.com/google/uuid"
	"io"
)

//go:generate mockgen -source=storage.go -destination=../storage/mocks/storage.go -package=mocks
type Storager interface {
	SaveURL(context.Context, models.ShortenURL) (models.ShortURL, error)
	GetURL(context.Context, models.ShortURL) (models.OriginalURL, error)
	GetShortURL(context.Context, models.OriginalURL) (models.ShortURL, error)
	InsertBatch(context.Context, []models.BatchRequest, models.Host, uuid.UUID) ([]models.BatchResponse, error)
	ListByUserID(context.Context, uuid.UUID) ([]models.ShortenURL, error)
	io.Closer
}

func GetStorage(flags config.Config) (Storager, error) {
	var db Storager
	var err error

	if flags.ConnectionString != "" {
		db, err = postgresql.New(flags.ConnectionString)
		if err != nil {
			logger.Sugar.Infow("Postgresql storage init error.")
			return nil, err
		}
	} else if flags.FileStoragePath != "" {
		db, err = file.New(flags.FileStoragePath)
		if err != nil {
			logger.Sugar.Infow("File storage init error.")
			return nil, err
		}
	} else {
		db = memory.New()
	}
	return db, nil
}
