package storage

import (
	"github.com/dubrovsky1/url-shortener/internal/config"
	"github.com/dubrovsky1/url-shortener/internal/storage/file"
	"github.com/dubrovsky1/url-shortener/internal/storage/memory"
	"github.com/dubrovsky1/url-shortener/internal/storage/postgresql"
	"github.com/dubrovsky1/url-shortener/internal/storage/repository"
	"log"
)

func GetStorage(flags config.Config) repository.Repository {
	var db repository.Repository
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
