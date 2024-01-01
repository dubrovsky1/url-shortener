package handlers

import (
	"github.com/dubrovsky1/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"

	"io"
	"log"
	"net/http"
	"net/url"
)

type Handler struct {
	Urls           storage.Storage
	ResultShortURL string
}

func New(s string, db *storage.Storage) *Handler {
	return &Handler{
		Urls:           *db,
		ResultShortURL: s,
	}
}

func (h Handler) GetURL(res http.ResponseWriter, req *http.Request) {
	shortURL := chi.URLParam(req, "id")
	log.Printf("Request Log. shortURL: %s\n", shortURL)

	originalURL, err := h.Urls.Get(shortURL)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("content-type", "text/plain")
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)

	log.Printf("Response Log. content-type: %s, Location: %s\n", res.Header().Get("content-type"), res.Header().Get("Location"))
}

func (h Handler) SaveURL(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	log.Printf("Request Log. Body: %s\n", body)

	//проверяем корректность url из тела запроса
	if err != nil || len(body) == 0 {
		http.Error(res, "The request body is missing", http.StatusBadRequest)
		return
	}
	if _, errParseURL := url.Parse(string(body)); errParseURL != nil {
		http.Error(res, "Not valid result URL", http.StatusBadRequest)
		return
	}

	//сохраняем в базу
	shortURL, errSave := h.Urls.Save(string(body))
	if errSave != nil {
		http.Error(res, "Save shortURL error", http.StatusBadRequest)
		return
	}

	//формируем тело ответа и проверяем на валидность
	var responseBody string
	if h.ResultShortURL == req.Host {
		responseBody = h.ResultShortURL + shortURL
	} else {
		responseBody = "http://" + req.Host + "/" + shortURL
	}

	if _, e := url.Parse(responseBody); e != nil {
		http.Error(res, "Not valid result URL", http.StatusBadRequest)
		return
	}

	res.Header().Set("content-type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	io.WriteString(res, responseBody)

	log.Printf("Response Log. content-type: %s, shortURL: %s\n", res.Header().Get("content-type"), shortURL)
}
