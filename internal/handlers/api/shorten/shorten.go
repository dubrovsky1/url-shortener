package shorten

import (
	"encoding/json"
	"errors"
	errs "github.com/dubrovsky1/url-shortener/internal/errors"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/dubrovsky1/url-shortener/internal/service"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
)

func Shorten(s *service.Service, resultShortURL string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		userID := ctx.Value(models.KeyUserID("UserID")).(uuid.UUID)
		body, err := io.ReadAll(req.Body)

		//logger.Sugar.Infow("Request shorten Log.", "Body", string(body), "userID", userID)

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

		res.Header().Set("content-type", "application/json")

		item := models.ShortenURL{
			OriginalURL: models.OriginalURL(r.URL),
			UserID:      userID,
		}

		//сохраняем в базу
		shortURL, errSave := s.SaveURL(ctx, item)
		if errSave != nil && !errors.Is(errSave, errs.ErrUniqueIndex) {
			http.Error(res, "Save shortURL error", http.StatusBadRequest)
			return
		}

		//если сохраняемый URL уже есть в базе, также формируем и возвращаем его короткую ссылку, но со статусом 409
		if errors.Is(errSave, errs.ErrUniqueIndex) {
			res.WriteHeader(http.StatusConflict)
		} else {
			res.WriteHeader(http.StatusCreated)
		}

		//формируем тело ответа и проверяем на валидность
		var responseURL string
		if resultShortURL == req.Host {
			responseURL = resultShortURL + string(shortURL)
		} else {
			responseURL = "http://" + req.Host + "/" + string(shortURL)
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

		res.Write(resp)

		//logger.Sugar.Infow(
		//	"Response Log.",
		//	"content-type", res.Header().Get("content-type"),
		//	"shortURL", shortURL,
		//	"URL from body json", responseURL,
		//)
	}
}
