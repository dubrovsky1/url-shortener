package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dubrovsky1/url-shortener/internal/generator"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"time"
)

type Storage struct {
	DB *sql.DB
}

func New(connectString string) (*Storage, error) {
	db, err := sql.Open("pgx", connectString)
	if err != nil {
		log.Fatal("Postgresql New. Database connection error. ", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		log.Fatal("Postgresql New. PingContext error. ", err)
	}

	queryString := `
						create table if not exists shortenUrls
						(
							id serial primary key,
							url text not null,
							shortenUrl text not null unique
						);
	
						comment on table shortenUrls is 'Таблица для сокращенных ссылок';
	
						comment on column shortenUrls.id is 'Идентификатор';
						comment on column shortenUrls.url is 'Оригинальный URL';
						comment on column shortenUrls.shortenUrl is 'Сокращенный URL';
					`

	_, err = db.ExecContext(ctx, queryString)
	if err != nil {
		log.Fatal("Postgresql New. Init query exec error. ", err)
	}

	return &Storage{DB: db}, nil
}

func (s *Storage) Save(originalURL string) (string, error) {
	//гененрируем короткую ссылку
	shortURL := generator.GetShortURL()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.DB.ExecContext(ctx, `insert into shortenUrls (url, shortenUrl) select $1 as url, $2 as shortenUrl;`, originalURL, shortURL)
	if err != nil {
		log.Fatal("Postgresql Save. Insert error. ", err)
	}

	return shortURL, nil
}

func (s *Storage) Get(shortURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var originalURL string

	row := s.DB.QueryRowContext(ctx, `select s.url from shortenUrls s where s.shortenUrl = $1;`, shortURL)

	err := row.Scan(&originalURL)
	if err != nil {
		return "", errors.New("Scan error")
	}

	return originalURL, nil
}
