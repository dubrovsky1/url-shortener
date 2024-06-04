package geturl

import (
	"errors"
	"github.com/dubrovsky1/url-shortener/internal/middleware/gzip"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/dubrovsky1/url-shortener/internal/service"
	"github.com/dubrovsky1/url-shortener/internal/storage/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetURL(t *testing.T) {
	logger.Initialize()

	tests := []models.TestCase{
		{
			Name: "Get url. Success.",
			Ms: models.MockStorage{
				Ctrl:       gomock.NewController(t),
				ShortURL:   "4fafrx",
				ShortenURL: models.ShortenURL{OriginalURL: "https://practicum.yandex.ru/", IsDel: false},
				Error:      nil,
			},
			Rp: models.RequestParams{
				Method: http.MethodGet,
				Body:   "",
			},
			Want: models.Want{
				ExpectedCode:        http.StatusTemporaryRedirect,
				ExpectedContentType: "text/plain",
				ExpectedLocation:    "https://practicum.yandex.ru/",
			},
		},
		{
			Name: "Get. Not exists short url.",
			Ms: models.MockStorage{
				Ctrl:       gomock.NewController(t),
				ShortURL:   "abcdef",
				ShortenURL: models.ShortenURL{},
				Error:      errors.New("the short url is missing"),
			},
			Rp: models.RequestParams{
				Method: http.MethodGet,
				Body:   "",
			},
			Want: models.Want{
				ExpectedCode: http.StatusBadRequest,
			},
		},
		{
			Name: "Get. Deleted url.",
			Ms: models.MockStorage{
				Ctrl:       gomock.NewController(t),
				ShortURL:   "4fafrx",
				ShortenURL: models.ShortenURL{OriginalURL: "https://practicum.yandex.ru/", IsDel: true},
				Error:      nil,
			},
			Rp: models.RequestParams{
				Method: http.MethodGet,
				Body:   "",
			},
			Want: models.Want{
				ExpectedCode: http.StatusGone,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			//хранилище-заглушка
			defer tt.Ms.Ctrl.Finish()

			storage := mocks.NewMockStorager(tt.Ms.Ctrl)
			serv := service.New(storage, 10, 10*time.Second)

			storage.EXPECT().GetURL(gomock.Any(), tt.Ms.ShortURL).Return(tt.Ms.ShortenURL, tt.Ms.Error)

			//маршрутизация запроса
			r := chi.NewRouter()
			r.Get("/{id}", logger.WithLogging(gzip.GzipMiddleware(GetURL(serv))))

			//создание http сервера
			ts := httptest.NewServer(r)
			defer ts.Close()

			URL := ts.URL + "/" + string(tt.Ms.ShortURL)

			req, errReq := http.NewRequest(tt.Rp.Method, URL, strings.NewReader(tt.Rp.Body))
			require.NoError(t, errReq)

			//запрет редиректа
			client := ts.Client()
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}

			resp, errResp := client.Do(req)
			require.NoError(t, errResp)

			defer resp.Body.Close()

			assert.Equal(t, tt.Want.ExpectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

			if tt.Want.ExpectedCode != http.StatusBadRequest && tt.Want.ExpectedCode != http.StatusGone {
				assert.Equal(t, tt.Want.ExpectedContentType, resp.Header.Get("content-type"), "content-type не совпадает с ожидаемым")
				assert.Equal(t, string(tt.Ms.ShortenURL.OriginalURL), resp.Header.Get("Location"), "Location не совпадает с ожидаемым")
			}

			t.Log("=============================================================>")
		})
	}
}

func BenchmarkGetURL(b *testing.B) {
	//logger.Initialize()
	ctrl := gomock.NewController(b)
	storage := mocks.NewMockStorager(ctrl)
	res := models.ShortenURL{OriginalURL: "https://practicum.yandex.ru/", IsDel: false}

	storage.EXPECT().GetURL(gomock.Any(), models.ShortURL("4fafrx")).Return(res, nil).AnyTimes()

	serv := service.New(storage, 10, 10*time.Second)

	//маршрутизация запроса
	r := chi.NewRouter()
	r.Get("/{id}", gzip.GzipMiddleware(GetURL(serv)))

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/4fafrx", strings.NewReader(""))
	rec := httptest.NewRecorder()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rec, req)
	}
}
