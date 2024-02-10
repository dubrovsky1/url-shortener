package storage

import (
	"github.com/dubrovsky1/url-shortener/internal/config"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/storage/file"
	"github.com/dubrovsky1/url-shortener/internal/storage/memory"
	"github.com/dubrovsky1/url-shortener/internal/storage/postgresql"
	"github.com/dubrovsky1/url-shortener/internal/storage/repository"
)

func GetStorage(flags config.Config) (repository.Repository, error) {
	var db repository.Repository
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
