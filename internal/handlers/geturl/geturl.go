package geturl

import (
	"context"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/go-chi/chi/v5"
	"net/http"
)

//go:generate mockgen -source=geturl.go -destination=../mocks/geturl.go -package=mocks
type URLGetter interface {
	GetURL(context.Context, string) (string, error)
}

func GetURL(db URLGetter) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		shortURL := chi.URLParam(req, "id")
		logger.Sugar.Infow("Request Log.", "shortURL", shortURL)

		originalURL, err := db.GetURL(ctx, shortURL)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		res.Header().Set("content-type", "text/plain")
		res.Header().Set("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)

		logger.Sugar.Infow(
			"Response Log.",
			"content-type", res.Header().Get("content-type"),
			"Location", res.Header().Get("Location"),
		)
	}
}
