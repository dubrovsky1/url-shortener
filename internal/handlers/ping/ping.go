package ping

import (
	"github.com/dubrovsky1/url-shortener/internal/storage/postgresql"
	"net/http"
)

func Ping(connectionString string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		db, err := postgresql.New(connectionString)
		if err != nil {
			http.Error(res, "database connection error", http.StatusInternalServerError)
			return
		}
		defer db.DB.Close()

		res.Header().Set("content-type", "text/plain")
		res.WriteHeader(http.StatusOK)
	}
}
