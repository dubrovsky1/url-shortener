package handlers

import (
	"encoding/json"
	"github.com/dubrovsky1/url-shortener/internal/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/dubrovsky1/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"

	"io"
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
	logger.Sugar.Infow("Request Log.", "shortURL", shortURL)

	originalURL, err := h.Urls.Get(shortURL)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("content-type", "text/plain")
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)

	logger.Sugar.Infow(
		"Request Log.",
		"content-type", res.Header().Get("content-type"),
		"Location", res.Header().Get("Location"),
	)
}

func (h Handler) SaveURL(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	logger.Sugar.Infow("Response Log.", "Body", string(body))

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

	logger.Sugar.Infow(
		"Response Log.",
		"content-type", res.Header().Get("content-type"),
		"shortURL", shortURL,
	)
}

func (h Handler) Shorten(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	logger.Sugar.Infow("Response Shorten Log.", "Body", string(body))

	//проверяем корректность url из тела запроса
	if err != nil || len(body) == 0 {
		http.Error(res, "The request body is missing", http.StatusBadRequest)
		return
	}

	//в теле запроса получили json со ссылкой - десериализируем его в объект запроса
	var r models.Request

	if err = json.Unmarshal(body, &r); err != nil {
		http.Error(res, "Bad json", http.StatusBadRequest)
		return
	}

	if _, errParseURL := url.Parse(r.URL); errParseURL != nil {
		http.Error(res, "Not valid result URL", http.StatusBadRequest)
		return
	}

	//сохраняем в базу
	shortURL, errSave := h.Urls.Save(r.URL)
	if errSave != nil {
		http.Error(res, "Save shortURL error", http.StatusBadRequest)
		return
	}

	//формируем тело ответа и проверяем на валидность
	var responseURL string
	if h.ResultShortURL == req.Host {
		responseURL = h.ResultShortURL + shortURL
	} else {
		responseURL = "http://" + req.Host + "/" + shortURL
	}

	if _, e := url.Parse(responseURL); e != nil {
		http.Error(res, "Not valid result URL", http.StatusBadRequest)
		return
	}

	//создаем объект ответа models.Response и сериализуем его в json resp, который возвращаем в теле ответа
	resp, err := json.Marshal(models.Response{Result: responseURL})

	if err != nil {
		http.Error(res, "resp marshal error", http.StatusBadRequest)
		return
	}

	res.Header().Set("content-type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(resp)

	logger.Sugar.Infow(
		"Response Log.",
		"content-type", res.Header().Get("content-type"),
		"shortURL", shortURL,
	)
}
