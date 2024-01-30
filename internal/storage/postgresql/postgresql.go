package postgresql

import (
	"context"
	"database/sql"
	"errors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
)

type Storage struct {
	DB *sql.DB
}

func New(connectString string) (*Storage, error) {
	db, err := sql.Open("pgx", connectString)
	if err != nil {
		return nil, errors.New("DataBase connection error")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, errors.New("DataBase connection error")
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
		return nil, errors.New("Init query exec error")
	}

	return &Storage{DB: db}, nil
}
