package shorten

import (
	"encoding/json"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"io"
	"net/http"
	"net/url"
)

//go:generate mockgen -source=saveurl.go -destination=../mocks/saveurl.go -package=mocks
type URLSaver interface {
	Save(string) (string, error)
}

func Shorten(db URLSaver, resultShortURL string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		logger.Sugar.Infow("Response Shorten Log.", "Body", string(body))

		//проверяем корректность url из тела запроса
		if err != nil {
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
		shortURL, errSave := db.Save(r.URL)
		if errSave != nil {
			http.Error(res, "Save shortURL error", http.StatusBadRequest)
			return
		}

		//формируем тело ответа и проверяем на валидность
		var responseURL string
		if resultShortURL == req.Host {
			responseURL = resultShortURL + shortURL
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
}
