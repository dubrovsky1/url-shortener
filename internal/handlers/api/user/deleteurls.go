package user

import (
	"encoding/json"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/dubrovsky1/url-shortener/internal/service"
	"github.com/google/uuid"
	"io"
	"net/http"
)

func DeleteURL(s *service.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		userID := ctx.Value(models.KeyUserID("UserID")).(uuid.UUID)
		body, err := io.ReadAll(req.Body)

		logger.Sugar.Infow("Request Log.", "Body", string(body), "userID", userID)

		if err != nil {
			http.Error(res, "The request body is missing", http.StatusBadRequest)
			return
		}

		var data []models.ShortURL

		if err = json.Unmarshal(body, &data); err != nil {
			http.Error(res, "Bad json", http.StatusBadRequest)
			return
		}

		deletedItems := make([]models.DeletedURLS, len(data))

		for i, item := range data {
			deletedItems[i] = models.DeletedURLS{ShortURL: item, UserID: userID}
		}

		//logger.Sugar.Infow("DeleteURL handler log.", "data", data, "deletedItems", deletedItems)

		err = s.DeleteURL(ctx, deletedItems)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		res.Header().Set("content-type", "text/plain")
		res.WriteHeader(http.StatusAccepted)
	}
}
