package postgresql

import (
	"context"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/google/uuid"
	"net/url"
)

func (s *Storage) GetURL(ctx context.Context, shortURL models.ShortURL) (models.ShortenURL, error) {
	var shortenURL models.ShortenURL

	row := s.DB.QueryRowContext(ctx, `
												select s.original_url,
												       s.is_deleted
												from shorten_urls s 
												where s.shorten_url = $1;
		`, shortURL,
	)

	err := row.Scan(&shortenURL.OriginalURL, &shortenURL.IsDel)
	if err != nil {
		logger.Sugar.Infow("Postgresql GetURL. Scan error.", "error", err.Error())
		return models.ShortenURL{}, err
	}
	return shortenURL, nil
}

func (s *Storage) GetShortURL(ctx context.Context, originalURL models.OriginalURL) (models.ShortURL, error) {
	var shortURL models.ShortURL

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

func (s *Storage) ListByUserID(ctx context.Context, host models.Host, u uuid.UUID) ([]models.ShortenURL, error) {
	var result []models.ShortenURL

	rows, err := s.DB.QueryContext(ctx, `
												select s.original_url,
												       s.shorten_url
												from shorten_urls s 
												where s.created_user_id = $1;
		`, u,
	)
	if err != nil {
		logger.Sugar.Infow("Postgresql GetByUserId. QueryContext error.")
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cur models.ShortenURL

		err = rows.Scan(&cur.OriginalURL, &cur.ShortURL)
		if err != nil {
			logger.Sugar.Infow("Postgresql GetByUserId. Scan error.")
			return nil, err
		}

		//составляем результирующий сокращённый URL и добавляем в слайс
		resultShortURL := "http://" + string(host) + "/" + string(cur.ShortURL)

		if _, e := url.Parse(resultShortURL); e != nil {
			logger.Sugar.Infow("Postgresql InsertBatch. Not result URL.")
			return nil, e
		}

		cur.ShortURL = models.ShortURL(resultShortURL)
		result = append(result, cur)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}
