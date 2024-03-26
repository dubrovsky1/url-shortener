package file

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	errs "github.com/dubrovsky1/url-shortener/internal/errors"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/google/uuid"
	"net/url"
	"os"
	"path/filepath"
)

type ShortenURL struct {
	UUID        uint               `json:"uuid"`
	ShortURL    models.ShortURL    `json:"short_url"`
	OriginalURL models.OriginalURL `json:"original_url"`
	UserID      uuid.UUID          `json:"user_id"`
}

type Storage struct {
	Urls     []ShortenURL
	Filename string
	maxUUID  uint
}

func (s *Storage) Close() error {
	return nil
}

func New(filename string) (*Storage, error) {
	var s Storage
	s.Filename = filename
	s.maxUUID = 0

	dir := filepath.Dir(filename)

	//создаем папку, если ее нет
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0666)
		if err != nil {
			logger.Sugar.Infow("Create dir error.")
			return nil, err
		}
	}

	//создаем файл, если его нет
	if _, err := os.Stat(s.Filename); os.IsNotExist(err) {
		newFile, errCreate := os.Create(s.Filename)
		if errCreate != nil {
			logger.Sugar.Infow("Create file error.")
			return nil, errCreate
		}
		newFile.Close()
	}

	file, err := os.Open(s.Filename)
	if err != nil {
		logger.Sugar.Infow("Open file error.")
		return nil, err
	}
	defer file.Close()

	//сканнер для чтения данных из файла
	scanner := bufio.NewScanner(file)

	//читаем файл построчно и заполняем структуру только при инициализации хранилища, чтобы каждый раз не считывать данные из файла
	for scanner.Scan() {
		data := scanner.Bytes()
		currentShortenURL := ShortenURL{}

		err = json.Unmarshal(data, &currentShortenURL)

		if err != nil {
			logger.Sugar.Infow("Unmarshal currentShortenURL error.")
			return nil, err
		}
		s.Urls = append(s.Urls, currentShortenURL)

		s.maxUUID = currentShortenURL.UUID
	}

	return &s, nil
}

func (s *Storage) SaveURL(ctx context.Context, item models.ShortenURL) (models.ShortURL, error) {
	//поиск уже сохраненной оригинальной ссылки
	shortURL, err := s.GetShortURL(ctx, item.OriginalURL)
	if err != nil {
		return shortURL, err
	}

	//создаем объект с сокращенной ссылкой, добавляем в хранилище и записываем в конец файла
	su := ShortenURL{
		UUID:        s.maxUUID + 1,
		ShortURL:    item.ShortURL,
		OriginalURL: item.OriginalURL,
		UserID:      item.UserID,
	}

	s.Urls = append(s.Urls, su)
	s.maxUUID++

	data, err := json.Marshal(&su)
	if err != nil {
		logger.Sugar.Infow("Marshal su error.")
		return "", err
	}

	//открываем файл, чтобы начать с ним работать
	file, err := os.OpenFile(s.Filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Sugar.Infow("Open file error.")
		return "", err
	}
	defer file.Close()

	//поле для записи в файл
	writer := bufio.NewWriter(file)

	// записываем событие в буфер
	if _, err = writer.Write(data); err != nil {
		logger.Sugar.Infow("Write file error.")
		return "", err
	}

	// добавляем перенос строки
	if err = writer.WriteByte('\n'); err != nil {
		logger.Sugar.Infow("Write \\n error.")
		return "", err
	}

	// записываем буфер в файл
	writer.Flush()

	return item.ShortURL, nil
}

func (s *Storage) GetURL(ctx context.Context, shortURL models.ShortURL) (models.OriginalURL, error) {
	//делаю поиск по массиву из Storage, тк чтение из файла происходит при инициализации хранилища
	for _, r := range s.Urls {
		if r.ShortURL == shortURL {
			return r.OriginalURL, nil
		}
	}
	return "", errors.New("the short url is missing")
}

func (s *Storage) GetShortURL(ctx context.Context, originalURL models.OriginalURL) (models.ShortURL, error) {
	for _, r := range s.Urls {
		if r.OriginalURL == originalURL {
			return r.ShortURL, errs.ErrUniqueIndex
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
			curItem.ShortURL, err = s.SaveURL(ctx, curItem)
			if err != nil {
				logger.Sugar.Infow("File InsertBatch. Insert error.")
				return nil, err
			}
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

func (s *Storage) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.ShortenURL, error) {
	var result []models.ShortenURL

	for _, row := range s.Urls {
		if row.UserID == userID {
			var curItem = models.ShortenURL{
				OriginalURL: row.OriginalURL,
				ShortURL:    row.ShortURL,
			}
			result = append(result, curItem)
		}
	}
	return result, nil
}
