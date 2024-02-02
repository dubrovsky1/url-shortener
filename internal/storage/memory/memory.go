package memory

import (
	"context"
	"errors"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"net/url"

	"github.com/dubrovsky1/url-shortener/internal/generator"
)

type Storage struct {
	urls map[string]string
}

func New() *Storage {
	return &Storage{make(map[string]string)}
}

func (s *Storage) SaveURL(ctx context.Context, originalURL string) (string, error) {
	//гененрируем короткую ссылку
	shortURL := generator.GetShortURL()

	//запоминаем url, соответствующий короткой ссылке
	s.urls[shortURL] = originalURL

	return shortURL, nil
}

func (s *Storage) GetURL(ctx context.Context, shortURL string) (string, error) {
	if _, ok := s.urls[shortURL]; !ok {
		return "", errors.New("the short url is missing")
	}
	return s.urls[shortURL], nil
}

func (s *Storage) InsertBatch(ctx context.Context, batch []models.BatchRequest, host string) ([]models.BatchResponse, error) {
	var result []models.BatchResponse

	for _, row := range batch {
		//гененрируем короткую ссылку
		shortURL := generator.GetShortURL()

		//запоминаем url, соответствующий короткой ссылке
		s.urls[shortURL] = row.URL

		//составляем результирующий сокращённый URL и добавляем в массив
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

	return result, nil
}
