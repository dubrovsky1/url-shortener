package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dubrovsky1/url-shortener/internal/generator"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/dubrovsky1/url-shortener/internal/storage/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"net/url"
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
                        create table if not exists shorten_urls
                        (
                            id             serial primary key,
                            original_url   text   not null,
                            shorten_url    text   not null unique
                        );
                        
                        comment on table shorten_urls is 'Таблица для сокращенных ссылок';
                        
                        comment on column shorten_urls.id is 'Идентификатор';
                        comment on column shorten_urls.original_url is 'Оригинальный URL';
                        comment on column shorten_urls.shorten_url is 'Сокращенный URL';
                                
                        create unique index if not exists uix_original_url on shorten_urls (original_url);
					`

	_, err = db.ExecContext(ctx, queryString)
	if err != nil {
		log.Fatal("Postgresql New. Init query exec error. ", err)
	}

	return &Storage{DB: db}, nil
}

func (s *Storage) SaveURL(ctx context.Context, originalURL string) (string, error) {
	//гененрируем короткую ссылку
	shortURL := generator.GetShortURL()

	_, err := s.DB.ExecContext(ctx, `
												insert into shorten_urls 
												(
													original_url, 
													shorten_url
												) 
												select $1 as original_url, 
												       $2 as shorten_url;
		`, originalURL, shortURL,
	)

	if err != nil {
		//проверка на ошибку вставки при нарушении уникальности индекса по оригинальным ссылкам
		var pgErr *pgconn.PgError

		//As - попытка привести возникшую при запросе ошибку err к "ошибкам в базах postgres"
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			err = repository.ErrUniqueIndex

			//поиск короткой ссылки по уже сохраненному в бд оригинальному URL
			var errGetShortURL error
			shortURL, errGetShortURL = s.GetShortURL(ctx, originalURL)
			if errGetShortURL != nil {
				logger.Sugar.Infow("Postgresql SaveURL. Find ShortURL error.")
				return "", errGetShortURL
			}
			return shortURL, err
		}
		//случай, если возникла ошибка, которая не связана с дублированием originalURL
		logger.Sugar.Infow("Postgresql SaveURL. Insert error.")
		return "", err
	}

	return shortURL, nil
}

func (s *Storage) GetURL(ctx context.Context, shortURL string) (string, error) {
	var originalURL string

	row := s.DB.QueryRowContext(ctx, `
												select s.original_url 
												from shorten_urls s 
												where s.shorten_url = $1;
		`, shortURL,
	)

	err := row.Scan(&originalURL)
	if err != nil {
		logger.Sugar.Infow("Postgresql GetURL. Scan error.")
		return "", err
	}
	return originalURL, nil
}

func (s *Storage) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	var shortURL string

	row := s.DB.QueryRowContext(ctx, `
												select s.shorten_url 
												from shorten_urls s 
												where s.original_url = $1;
		`, originalURL,
	)

	err := row.Scan(&shortURL)
	if err != nil {
		logger.Sugar.Infow("Postgresql GetShortURL. Scan error.")
		return "", err
	}

	return shortURL, nil
}

func (s *Storage) InsertBatch(ctx context.Context, batch []models.BatchRequest, host string) ([]models.BatchResponse, error) {
	var result []models.BatchResponse

	//открытие транзакции
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		logger.Sugar.Infow("Postgresql InsertBatch. Begin transaction error.")
		return nil, err
	}
	// если Commit будет раньше, то откат проигнорируется
	defer tx.Rollback()

	//подготовка скомпилированных запросов
	insertQuery, err := tx.PrepareContext(ctx, `
                                                   insert into shorten_urls 
                                                   (
                                                       original_url, 
                                                       shorten_url
                                                   ) 
                                                   select $1 as original_url, 
                                                          $2 as shorten_url
                                                   on conflict (original_url) 
                                                   do nothing;
	`)
	if err != nil {
		logger.Sugar.Infow("Postgresql InsertBatch. Prepare query insert error.")
		return nil, err
	}
	defer insertQuery.Close()

	selectShortURLQuery, err := tx.PrepareContext(ctx, `
                                                                 select su.shorten_url
                                                                 from shorten_urls su
                                                                 where su.original_url = $1;
	`)
	if err != nil {
		logger.Sugar.Infow("Postgresql InsertBatch. Prepare query select shorten_url error.")
		return nil, err
	}
	defer selectShortURLQuery.Close()

	for _, row := range batch {
		//гененрируем короткую ссылку
		shortURL := generator.GetShortURL()

		//прикрепляем к транзакции выполнение запроса поиска shorten_url, если original_url уже есть в базе
		res := selectShortURLQuery.QueryRowContext(ctx, row.URL)
		err = res.Scan(&shortURL)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.Sugar.Infow("Postgresql InsertBatch. Scan error.")
			return nil, err
		}

		//прикрепляем к транзакции выполнение запроса вставки, передавая в скомпилированный запрос данные по каждой ссылке из входящего слайса
		_, err = insertQuery.ExecContext(ctx, row.URL, shortURL)
		if err != nil {
			logger.Sugar.Infow("Postgresql InsertBatch. ExecContext error.")
			return nil, err
		}

		//составляем результирующий сокращённый URL и добавляем в слайс
		resultShortURL := "http://" + host + "/" + shortURL

		if _, e := url.Parse(resultShortURL); e != nil {
			logger.Sugar.Infow("Postgresql InsertBatch. Not result URL.")
			return nil, e
		}

		r := models.BatchResponse{
			CorrelationID: row.CorrelationID,
			ShortURL:      resultShortURL,
		}

		result = append(result, r)
	}

	tx.Commit()

	return result, nil
}
