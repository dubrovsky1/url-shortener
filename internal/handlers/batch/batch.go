package batch

import (
	"context"
	"encoding/json"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"io"
	"net/http"
	"net/url"
)

//go:generate mockgen -source=batch.go -destination=../mocks/batch.go -package=mocks
type BatchURLSaver interface {
	InsertBatch(context.Context, []models.BatchRequest, string) ([]models.BatchResponse, error)
}

func Batch(db BatchURLSaver) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		body, err := io.ReadAll(req.Body)
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

		result, err := db.InsertBatch(ctx, data, req.Host)
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
