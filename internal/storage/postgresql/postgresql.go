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

	queryString := `
						create schema if not exists dbo;
	
						grant all privileges on schema dbo to sa;
	
						create table if not exists dbo.shortenUrls
						(
							id serial primary key,
							url text not null,
							shortenUrl text not null unique
						);
	
						comment on table dbo.shortenUrls is 'Таблица для сокращенных ссылок';
	
						comment on column dbo.shortenUrls.id is 'Идентификатор';
						comment on column dbo.shortenUrls.url is 'Оригинальный URL';
						comment on column dbo.shortenUrls.shortenUrl is 'Сокращенный URL';
					`

	_, err = db.Exec(queryString)
	if err != nil {
		log.Fatal("Postgresql New. Init query exec error. ", err)
	}

	return &Storage{DB: db}, nil
}
