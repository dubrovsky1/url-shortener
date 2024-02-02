package storage

import (
	"context"
	"github.com/dubrovsky1/url-shortener/internal/models"
)

type Repository interface {
	SaveURL(context.Context, string) (string, error)
	GetURL(context.Context, string) (string, error)
	InsertBatch(context.Context, []models.BatchRequest, string) ([]models.BatchResponse, error)
}
