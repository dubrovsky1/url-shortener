package memory

import (
	"context"
	"errors"
	errs "github.com/dubrovsky1/url-shortener/internal/errors"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/google/uuid"
	"net/url"

	"github.com/dubrovsky1/url-shortener/internal/generator"
)

type Storage struct {
	urls map[models.ShortURL]models.ShortenURL
}

func New() *Storage {
	return &Storage{make(map[models.ShortURL]models.ShortenURL)}
}

func (s *Storage) Close() error {
	return nil
}

func (s *Storage) SaveURL(ctx context.Context, item models.ShortenURL) (models.ShortURL, error) {
	//поиск уже сохраненной оригинальной ссылки
	shortURL, err := s.GetShortURL(ctx, item.OriginalURL)
	if err != nil {
		return shortURL, err
	}

	//запоминаем url, соответствующий короткой ссылке
	s.urls[item.ShortURL] = item

	return item.ShortURL, nil
}

func (s *Storage) GetURL(ctx context.Context, shortURL models.ShortURL) (models.ShortenURL, error) {
	if _, ok := s.urls[shortURL]; !ok {
		return models.ShortenURL{}, errors.New("the short url is missing")
	}
	return s.urls[shortURL], nil
}

func (s *Storage) GetShortURL(ctx context.Context, originalURL models.OriginalURL) (models.ShortURL, error) {
	for su, ou := range s.urls {
		if ou.OriginalURL == originalURL {
			return su, errs.ErrUniqueIndex
		}
	}
	return "", nil
}

func (s *Storage) InsertBatch(ctx context.Context, batch []models.BatchRequest, host models.Host, userID uuid.UUID) ([]models.BatchResponse, error) {
	var result []models.BatchResponse

	for _, row := range batch {
		var err error

		var curItem = models.ShortenURL{
			OriginalURL: models.OriginalURL(row.URL),
			UserID:      userID,
		}

		//поиск уже сохраненной оригинальной ссылки
		curItem.ShortURL, err = s.GetShortURL(ctx, curItem.OriginalURL)

		if err == nil {
			//гененрируем короткую ссылку
			curItem.ShortURL = models.ShortURL(generator.GetShortURL())

			//запоминаем url, соответствующий короткой ссылке
			s.urls[curItem.ShortURL] = curItem
		}

		//составляем результирующий сокращённый URL и добавляем в массив
		resultShortURL := "http://" + string(host) + "/" + string(curItem.ShortURL)

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

func (s *Storage) ListByUserID(ctx context.Context, host models.Host, userID uuid.UUID) ([]models.ShortenURL, error) {
	var result []models.ShortenURL

	for _, row := range s.urls {
		if row.UserID == userID {

			//составляем результирующий сокращённый URL и добавляем в массив
			resultShortURL := "http://" + string(host) + "/" + string(row.ShortURL)

			if _, e := url.Parse(resultShortURL); e != nil {
				logger.Sugar.Infow("Postgresql ListByUserID. Not result URL.")
				return nil, e
			}

			var curItem = models.ShortenURL{
				OriginalURL: row.OriginalURL,
				ShortURL:    models.ShortURL(resultShortURL),
			}
			result = append(result, curItem)
		}
	}
	return result, nil
}

func (s *Storage) DeleteURL(ctx context.Context, deletedItems []models.DeletedURLS) error {
	for _, item := range deletedItems {
		if _, ok := s.urls[item.ShortURL]; ok {
			deleted := models.ShortenURL{
				ID:          s.urls[item.ShortURL].ID,
				OriginalURL: s.urls[item.ShortURL].OriginalURL,
				ShortURL:    s.urls[item.ShortURL].ShortURL,
				UserID:      s.urls[item.ShortURL].UserID,
				IsDel:       true,
			}
			s.urls[item.ShortURL] = deleted
		}
	}
	return nil
}
