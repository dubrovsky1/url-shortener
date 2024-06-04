package saveurl

import (
	errs "github.com/dubrovsky1/url-shortener/internal/errors"
	"github.com/dubrovsky1/url-shortener/internal/middleware/auth"
	"github.com/dubrovsky1/url-shortener/internal/middleware/gzip"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/dubrovsky1/url-shortener/internal/service"
	"github.com/dubrovsky1/url-shortener/internal/storage/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSaveURL(t *testing.T) {
	logger.Initialize()

	tests := []models.TestCase{
		{
			Name: "Save url. Success.",
			Ms: models.MockStorage{
				Ctrl:        gomock.NewController(t),
				OriginalURL: "https://practicum.yandex.ru/",
				ShortURL:    "jB9Wbk",
				Error:       nil,
			},
			Rp: models.RequestParams{
				Method: http.MethodPost,
				Body:   "https://practicum.yandex.ru/",
			},
			Want: models.Want{
				ExpectedCode:        http.StatusCreated,
				ExpectedContentType: "text/plain",
				ExpectedShortURL:    "jB9Wbk",
			},
		},
		{
			Name: "Save url. Unique URL conflict.",
			Ms: models.MockStorage{
				Ctrl:        gomock.NewController(t),
				OriginalURL: "https://yandex.ru/",
				ShortURL:    "2Yy05g",
				Error:       errs.ErrUniqueIndex,
			},
			Rp: models.RequestParams{
				Method: http.MethodPost,
				Body:   "https://yandex.ru/",
			},
			Want: models.Want{
				ExpectedCode:        http.StatusConflict,
				ExpectedContentType: "text/plain",
				ExpectedShortURL:    "2Yy05g",
			},
		},
		{
			Name: "Save url. No exists body.",
			Ms: models.MockStorage{
				Ctrl:        gomock.NewController(t),
				OriginalURL: "https://yandex.ru/",
				ShortURL:    "2Yy05g",
				Error:       nil,
			},
			Rp: models.RequestParams{
				Method: http.MethodPost,
				Body:   "",
			},
			Want: models.Want{
				ExpectedCode: http.StatusBadRequest,
			},
		},
		{
			Name: "Save url. Not valid body original url.",
			Ms: models.MockStorage{
				Ctrl:        gomock.NewController(t),
				OriginalURL: "https://yandex.ru/",
				ShortURL:    "2Yy05g",
				Error:       nil,
			},
			Rp: models.RequestParams{
				Method: http.MethodPost,
				Body:   "sdaff/sde8%%%4325sa@.ru-213",
			},
			Want: models.Want{
				ExpectedCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			defer tt.Ms.Ctrl.Finish()

			storage := mocks.NewMockStorager(tt.Ms.Ctrl)
			serv := service.New(storage, 10, 10*time.Second)

			storage.EXPECT().SaveURL(gomock.Any(), gomock.Any()).Return(tt.Ms.ShortURL, tt.Ms.Error).AnyTimes()

			r := chi.NewRouter()
			r.Post("/", auth.Auth(logger.WithLogging(gzip.GzipMiddleware(SaveURL(serv, "http://localhost:8080/")))))

			ts := httptest.NewServer(r)
			defer ts.Close()

			URL := ts.URL + "/"

			req, errReq := http.NewRequest(tt.Rp.Method, URL, strings.NewReader(tt.Rp.Body))
			require.NoError(t, errReq)

			client := ts.Client()
			resp, errResp := client.Do(req)
			require.NoError(t, errResp)

			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.Want.ExpectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

			if tt.Want.ExpectedCode != http.StatusBadRequest {
				assert.Equal(t, tt.Want.ExpectedContentType, resp.Header.Get("content-type"), "content-type не совпадает с ожидаемым")
				assert.Equal(t, ts.URL+"/"+tt.Want.ExpectedShortURL, string(respBody), "Body не совпадает с ожидаемым")
			}

			t.Log("=============================================================>")
		})
	}
}

func BenchmarkSaveURL(b *testing.B) {
	ctrl := gomock.NewController(b)
	storage := mocks.NewMockStorager(ctrl)
	//res := models.ShortenURL{OriginalURL: "https://practicum.yandex.ru/", IsDel: false}

	storage.EXPECT().SaveURL(gomock.Any(), gomock.Any()).Return(models.ShortURL("jB9Wbk"), nil).AnyTimes()

	serv := service.New(storage, 10, 10*time.Second)

	//маршрутизация запроса
	r := chi.NewRouter()
	r.Post("/", auth.Auth(gzip.GzipMiddleware(SaveURL(serv, "http://localhost:8080/"))))

	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/", strings.NewReader("https://practicum.yandex.ru/"))
	rec := httptest.NewRecorder()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rec, req)
	}
}
