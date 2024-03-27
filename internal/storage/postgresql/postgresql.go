package postgresql

import (
	"context"
	"database/sql"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
)

type Storage struct {
	DB *sql.DB
}

func (s *Storage) Close() error {
	s.DB.Close()
	return nil
}

func New(connectString string) (*Storage, error) {
	db, err := sql.Open("pgx", connectString)
	if err != nil {
		logger.Sugar.Infow("Postgresql New. Database connection error.")
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		logger.Sugar.Infow("Postgresql New. PingContext error.")
		return nil, err
	}

	queryString := `
                        create table if not exists shorten_urls
                        (
                            id              serial primary key,
                            original_url    text   not null,
                            shorten_url     text   not null unique,
                            created_user_id uuid   null
                        );
                        
                        comment on table shorten_urls is 'Таблица для сокращенных ссылок';
                        
                        comment on column shorten_urls.id is 'Идентификатор';
                        comment on column shorten_urls.original_url is 'Оригинальный URL';
                        comment on column shorten_urls.shorten_url is 'Сокращенный URL';
                        comment on column shorten_urls.created_user_id is 'Id создавшего пользователя';
                                
                        create unique index if not exists uix_original_url on shorten_urls (original_url);
					`

	_, err = db.ExecContext(ctx, queryString)
	if err != nil {
		logger.Sugar.Infow("Postgresql New. Init query exec error.")
		return nil, err
	}

	return &Storage{DB: db}, nil
}
