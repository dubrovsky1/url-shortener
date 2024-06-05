package ping

import (
	"github.com/dubrovsky1/url-shortener/internal/middleware/gzip"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func BenchmarkPing(b *testing.B) {
	r := chi.NewRouter()
	r.Get("/ping", gzip.GzipMiddleware(Ping("")))

	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/ping", strings.NewReader(""))
	rec := httptest.NewRecorder()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rec, req)
	}
}
