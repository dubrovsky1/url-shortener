package shorten

import (
	"encoding/json"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/dubrovsky1/url-shortener/internal/service"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
)

func Batch(s *service.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		userID := ctx.Value("UserID").(uuid.UUID)
		body, err := io.ReadAll(req.Body)

		logger.Sugar.Infow("Request batch Log.", "Body", string(body), "userID", userID)

		if err != nil {
			http.Error(res, "The request body is missing", http.StatusBadRequest)
			return
		}

		//в теле запроса получили json со ссылкой - десериализируем его в объект запроса
		var data []models.BatchRequest

		if err = json.Unmarshal(body, &data); err != nil {
			http.Error(res, "Bad json", http.StatusBadRequest)
			return
		}

		for _, row := range data {
			logger.Sugar.Infow("Request body urls.",
				"correlation_id", row.CorrelationID,
				"original_url", row.URL)

			if _, errParseURL := url.Parse(row.URL); errParseURL != nil {
				http.Error(res, "Not valid original URL.", http.StatusBadRequest)
				return
			}
		}

		result, err := s.InsertBatch(ctx, data, models.Host(req.Host), userID)
		if err != nil {
			http.Error(res, "Insert error", http.StatusBadRequest)
			return
		}

		//создаем объект ответа models.Response и сериализуем его в json resp, который возвращаем в теле ответа
		resp, err := json.Marshal(result)
		if err != nil {
			http.Error(res, "resp marshal error", http.StatusBadRequest)
			return
		}

		res.Header().Set("content-type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(resp)

		for _, row := range result {
			logger.Sugar.Infow("Response result urls.",
				"correlation_id", row.CorrelationID,
				"original_url", row.ShortURL)
		}
	}
}
