package models

import "github.com/google/uuid"

type (
	OriginalURL string
	ShortURL    string
	Host        string
	KeyUserID   string
)

type ShortenURL struct {
	ID          uuid.UUID   `json:"-"`
	ShortURL    ShortURL    `json:"short_url,omitempty"`
	OriginalURL OriginalURL `json:"original_url,omitempty"`
	UserID      uuid.UUID   `json:"-"`
	IsDel       bool        `json:"is_deleted,omitempty"`
}

type DeletedURLS struct {
	UserID   uuid.UUID `db:"created_user_id"`
	ShortURL ShortURL  `db:"short_url"`
}
