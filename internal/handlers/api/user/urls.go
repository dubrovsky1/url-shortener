package user

import (
	"encoding/json"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/dubrovsky1/url-shortener/internal/service"
	"github.com/google/uuid"
	"net/http"
)

func ListByUserID(s *service.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		userID := ctx.Value(models.KeyUserID("UserID")).(uuid.UUID)
		logger.Sugar.Infow("Request Log.", "UserId", userID)

		result, err := s.ListByUserID(ctx, userID)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if len(result) == 0 {
			http.Error(res, "resp no content", http.StatusNoContent)
			return
		}

		//создаем объект ответа models.Response и сериализуем его в json resp, который возвращаем в теле ответа
		resp, err := json.Marshal(result)
		if err != nil {
			http.Error(res, "resp marshal error", http.StatusBadRequest)
			return
		}

		res.Header().Set("content-type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(resp)

		for _, row := range result {
			logger.Sugar.Infow("Response result urls.",
				"short_id", row.ShortURL,
				"original_url", row.OriginalURL)
		}
	}
}
