package service

import (
	"context"
	"github.com/dubrovsky1/url-shortener/internal/generator"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/google/uuid"
)

// для слоя бизнес логики не нужны реализации GetShortURL и io.Closer, поэтому их не включаем
//
//go:generate mockgen -source=service.go -destination=../service/mocks/service.go -package=mocks
type Storager interface {
	SaveURL(context.Context, models.ShortenURL) (models.ShortURL, error)
	GetURL(context.Context, models.ShortURL) (models.OriginalURL, error)
	InsertBatch(context.Context, []models.BatchRequest, models.Host, uuid.UUID) ([]models.BatchResponse, error)
	ListByUserId(context.Context, uuid.UUID) ([]models.ShortenURL, error)
}

type Service struct {
	storage Storager
}

func New(storage Storager) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) SaveURL(ctx context.Context, item models.ShortenURL) (models.ShortURL, error) {
	//гененрируем короткую ссылку
	item.ShortURL = models.ShortURL(generator.GetShortURL())

	shortURL, err := s.storage.SaveURL(ctx, item)
	if err != nil {
		return shortURL, err
	}
	return shortURL, nil
}

func (s *Service) GetURL(ctx context.Context, shortURL models.ShortURL) (models.OriginalURL, error) {
	originalURL, err := s.storage.GetURL(ctx, shortURL)
	if err != nil {
		return originalURL, err
	}
	return originalURL, nil
}

func (s *Service) InsertBatch(ctx context.Context, batch []models.BatchRequest, host models.Host, userID uuid.UUID) ([]models.BatchResponse, error) {
	result, err := s.storage.InsertBatch(ctx, batch, host, userID)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (s *Service) ListByUserId(ctx context.Context, userID uuid.UUID) ([]models.ShortenURL, error) {
	result, err := s.storage.ListByUserId(ctx, userID)
	if err != nil {
		return result, err
	}
	return result, nil
}
