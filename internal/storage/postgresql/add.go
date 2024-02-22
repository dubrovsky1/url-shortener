package postgresql

import (
	"context"
	"database/sql"
	"errors"
	errs "github.com/dubrovsky1/url-shortener/internal/errors"
	"github.com/dubrovsky1/url-shortener/internal/generator"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"net/url"
)

func (s *Storage) SaveURL(ctx context.Context, item models.ShortenURL) (models.ShortURL, error) {
	_, err := s.DB.ExecContext(ctx, `
												insert into shorten_urls 
												(
													original_url, 
													shorten_url,
												    created_user_id
												) 
												select $1 as original_url, 
												       $2 as shorten_url,
												       $3 as created_user_id;
		`, item.OriginalURL, item.ShortURL, item.UserID,
	)

	if err != nil {
		//проверка на ошибку вставки при нарушении уникальности индекса по оригинальным ссылкам
		var pgErr *pgconn.PgError

		//As - попытка привести возникшую при запросе ошибку err к "ошибкам в базах postgres"
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			err = errs.ErrUniqueIndex

			//поиск короткой ссылки по уже сохраненному в бд оригинальному URL
			shortURL, errGetShortURL := s.GetShortURL(ctx, item.OriginalURL)
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

	return item.ShortURL, nil
}

func (s *Storage) InsertBatch(ctx context.Context, batch []models.BatchRequest, host models.Host, userID uuid.UUID) ([]models.BatchResponse, error) {
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
                                                       shorten_url,
                                                       created_user_id
                                                   ) 
                                                   select $1 as original_url, 
                                                          $2 as shorten_url,
                                                          $3 as created_user_id
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
		_, err = insertQuery.ExecContext(ctx, row.URL, shortURL, userID)
		if err != nil {
			logger.Sugar.Infow("Postgresql InsertBatch. ExecContext error.")
			return nil, err
		}

		//составляем результирующий сокращённый URL и добавляем в слайс
		resultShortURL := "http://" + string(host) + "/" + shortURL

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
