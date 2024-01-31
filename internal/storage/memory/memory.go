package memory

import (
	"errors"

	"github.com/dubrovsky1/url-shortener/internal/generator"
)

type Storage struct {
	urls map[string]string
}

func New() *Storage {
	return &Storage{make(map[string]string)}
}

func (s *Storage) Save(originalURL string) (string, error) {
	//гененрируем короткую ссылку
	shortURL := generator.GetShortURL()

	//запоминаем url, соответствующий короткой ссылке
	s.urls[shortURL] = originalURL

	return shortURL, nil
}

func (s *Storage) Get(shortURL string) (string, error) {
	if _, ok := s.urls[shortURL]; !ok {
		return "", errors.New("the short url is missing")
	}
	return s.urls[shortURL], nil
}
