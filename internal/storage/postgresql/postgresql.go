package postgresql

import (
	"context"
	"database/sql"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"time"
)

type Storage struct {
	DB *sql.DB
}

func New(connectString string) (*Storage, error) {
	logger.Sugar.Infow("Postgresql New.", "connectString", connectString)

	db, err := sql.Open("pgx", connectString)
	if err != nil {
		log.Fatal("Postgresql New. Database connection error. ", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		log.Fatal("Postgresql New. PingContext error. ", err)
	}

	return &Storage{DB: db}, nil
}
