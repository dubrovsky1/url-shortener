package ping

import (
	"github.com/dubrovsky1/url-shortener/internal/config"
	"github.com/dubrovsky1/url-shortener/internal/middleware/gzip"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPing(t *testing.T) {
	logger.Initialize()
	flags := config.ParseFlags()

	tests := []models.RequestParams{
		{
			Name:             "Ping. Success.",
			Method:           http.MethodGet,
			ConnectionString: flags.ConnectionString,
			Want: models.Want{
				ExpectedCode: http.StatusOK,
			},
		},
		{
			Name:             "Ping. Fail.",
			Method:           http.MethodGet,
			ConnectionString: "host=localhost port=5432 user=sa password=123456 dbname=urls sslmode=disable",
			Want: models.Want{
				ExpectedCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Get("/ping", logger.WithLogging(gzip.GzipMiddleware(Ping(tt.ConnectionString))))
			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err_req := http.NewRequest(tt.Method, ts.URL+"/ping", strings.NewReader(""))
			require.NoError(t, err_req)

			client := ts.Client()
			resp, err_resp := client.Do(req)
			require.NoError(t, err_resp)

			defer resp.Body.Close()

			assert.Equal(t, tt.Want.ExpectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

			t.Log("=============================================================>")
		})
	}
}
