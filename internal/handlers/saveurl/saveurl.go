package saveurl

import (
	"errors"
	errs "github.com/dubrovsky1/url-shortener/internal/errors"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/dubrovsky1/url-shortener/internal/service"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
)

func SaveURL(s *service.Service, resultShortURL string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		//проверка наличия айди пользоователя в контексте
		//var userID uuid.UUID
		//val := ctx.Value("UserID").(uuid.UUID)
		//if val == uuid.Nil {
		//	userID = val.(uuid.UUID)
		//}

		userID := ctx.Value("UserID").(uuid.UUID)

		body, err := io.ReadAll(req.Body)

		logger.Sugar.Infow("Request Log.", "Body", string(body), "userID", userID)

		item := models.ShortenURL{
			OriginalURL: models.OriginalURL(body),
			UserID:      userID,
		}

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
		var responseBody string
		if resultShortURL == req.Host {
			responseBody = resultShortURL + string(shortURL)
		} else {
			responseBody = "http://" + req.Host + "/" + string(shortURL)
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
