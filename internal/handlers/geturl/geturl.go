package geturl

import (
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/dubrovsky1/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func GetURL(s *service.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		shortURL := models.ShortURL(chi.URLParam(req, "id"))
		logger.Sugar.Infow("Request Log.", "shortURL", shortURL)

		result, err := s.GetURL(ctx, shortURL)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if result.IsDel == true {
			http.Error(res, "deleted", http.StatusGone)
			return
		}

		res.Header().Set("content-type", "text/plain")
		res.Header().Set("Location", string(result.OriginalURL))
		res.WriteHeader(http.StatusTemporaryRedirect)

		logger.Sugar.Infow(
			"Response Log.",
			"content-type", res.Header().Get("content-type"),
			"Location", res.Header().Get("Location"),
		)
	}
}
