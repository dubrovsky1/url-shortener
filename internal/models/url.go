package models

import "github.com/google/uuid"

type (
	OriginalURL string
	ShortURL    string
	Host        string
)

type ShortenURL struct {
	ID          uuid.UUID   `json:"-"`
	ShortURL    ShortURL    `json:"short_url"`
	OriginalURL OriginalURL `json:"original_url"`
	UserID      uuid.UUID   `json:"-"`
}
