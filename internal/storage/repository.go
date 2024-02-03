package storage

import (
	"context"
	"errors"
	"github.com/dubrovsky1/url-shortener/internal/models"
)

var ErrUniqueIndex = errors.New("unique index error")

type Repository interface {
	SaveURL(context.Context, string) (string, error)
	GetURL(context.Context, string) (string, error)
	GetShortURL(context.Context, string) (string, error)
	InsertBatch(context.Context, []models.BatchRequest, string) ([]models.BatchResponse, error)
}
