package file

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/dubrovsky1/url-shortener/internal/generator"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"log"
	"os"
)

type ShortenURL struct {
	UUID        uint   `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Storage struct {
	Urls     []ShortenURL
	Filename string
	maxUUID  uint
}

func New(filename string) (*Storage, error) {
	var s Storage
	s.Filename = filename
	s.maxUUID = 0

	//создаем новый или присваиваем существующий файл хранилищу
	file, err := os.OpenFile(s.Filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
		}
		s.Urls = append(s.Urls, currentShortenURL)

		s.maxUUID = currentShortenURL.UUID
	}

	return &s, nil
}

func (s *Storage) Save(originalURL string) (string, error) {
	//гененрируем короткую ссылку
	shortURL := generator.GetShortURL()

	for _, r := range s.Urls {
		logger.Sugar.Infoln("Storage", "Urls", r)
	}

	//создаем объект с сокращенной ссылкой, добавляем в хранилище и записываем в конец файла
	su := ShortenURL{
		UUID:        s.maxUUID + 1,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	s.Urls = append(s.Urls, su)
	s.maxUUID++

	data, err := json.Marshal(&su)
	if err != nil {
		log.Fatal(err)
	}

	//открываем файл, чтобы начать с ним работать
	file, err := os.OpenFile(s.Filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	//поле для записи в файл
	writer := bufio.NewWriter(file)

	// записываем событие в буфер
	if _, err = writer.Write(data); err != nil {
		log.Fatal(err)
	}

	// добавляем перенос строки
	if err = writer.WriteByte('\n'); err != nil {
		log.Fatal(err)
	}

	// записываем буфер в файл
	writer.Flush()

	return shortURL, nil
}

func (s *Storage) Get(shortURL string) (string, error) {
	//делаю поиск по массиву из Storage, тк чтение из файла происходит при инициализации хранилища
	for _, r := range s.Urls {
		if r.ShortURL == shortURL {
			return r.OriginalURL, nil
		}
	}
	return "", errors.New("the short url is missing")
}
