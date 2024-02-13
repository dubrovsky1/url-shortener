package saveurl

import (
	"context"
	"errors"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/storage/repository"
	"io"
	"net/http"
	"net/url"
)

//go:generate mockgen -source=saveurl.go -destination=../mocks/saveurl.go -package=mocks
type URLSaver interface {
	SaveURL(context.Context, string) (string, error)
}

func SaveURL(db URLSaver, resultShortURL string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		body, err := io.ReadAll(req.Body)
		logger.Sugar.Infow("Request Log.", "Body", string(body))

		//проверяем корректность url из тела запроса
		if err != nil || len(body) == 0 {
			http.Error(res, "The request body is missing", http.StatusBadRequest)
			return
		}
		if _, errParseURL := url.Parse(string(body)); errParseURL != nil {
			http.Error(res, "Not valid result URL", http.StatusBadRequest)
			return
		}

		res.Header().Set("content-type", "text/plain")

		//сохраняем в базу
		shortURL, errSave := db.SaveURL(ctx, string(body))
		if errSave != nil && !errors.Is(errSave, repository.ErrUniqueIndex) {
			http.Error(res, "Save shortURL error", http.StatusBadRequest)
			return
		}

		//если сохраняемый URL уже есть в базе, также формируем и возвращаем его короткую ссылку, но со статусом 409
		if errors.Is(errSave, repository.ErrUniqueIndex) {
			res.WriteHeader(http.StatusConflict)
		} else {
			res.WriteHeader(http.StatusCreated)
		}

		//формируем тело ответа и проверяем на валидность
		var responseBody string
		if resultShortURL == req.Host {
			responseBody = resultShortURL + shortURL
		} else {
			responseBody = "http://" + req.Host + "/" + shortURL
		}

		if _, e := url.Parse(responseBody); e != nil {
			http.Error(res, "Not valid result URL", http.StatusBadRequest)
			return
		}

		io.WriteString(res, responseBody)

		logger.Sugar.Infow(
			"Response Log.",
			"content-type", res.Header().Get("content-type"),
			"shortURL", shortURL,
			"Body", responseBody,
		)
	}
}
