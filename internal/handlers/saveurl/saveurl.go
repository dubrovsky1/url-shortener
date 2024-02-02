package saveurl

import (
	"context"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
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

		//сохраняем в базу
		shortURL, errSave := db.SaveURL(ctx, string(body))
		if errSave != nil {
			http.Error(res, "Save shortURL error", http.StatusBadRequest)
			return
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

		res.Header().Set("content-type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		io.WriteString(res, responseBody)

		logger.Sugar.Infow(
			"Response Log.",
			"content-type", res.Header().Get("content-type"),
			"shortURL", shortURL,
		)
	}
}
