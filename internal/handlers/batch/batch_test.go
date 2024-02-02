package batch

import (
	"bytes"
	"github.com/dubrovsky1/url-shortener/internal/handlers/mocks"
	"github.com/dubrovsky1/url-shortener/internal/middleware/gzip"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//определяем хранилище-заглушку
	storage := mocks.NewMockBatchURLSaver(ctrl)

	//Передать в функцию можно что угодно - она должна это сохранить. Некорректрые url будут отсечены до сохранения в базу
	storage.EXPECT().InsertBatch(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	//создаем тестовый сервер, который будет проверять запросы, получаемые функцией-обработчиком хендлера batch
	logger.Initialize()
	r := chi.NewRouter()
	r.Post("/api/shorten/batch", logger.WithLogging(gzip.GzipMiddleware(Batch(storage))))
	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []models.RequestParams{
		{
			Name:   "Batch. Success.",
			Method: http.MethodPost,
			URL:    ts.URL + "/api/shorten/batch",
			JSONBody: bytes.NewBufferString(`
													[{
													    "correlation_id": "a",
													    "original_url": "https://123456.ru/"
													},
													{
													    "correlation_id": "b",
													    "original_url": "https://practicum.yandex.ru"
													},
													{
													    "correlation_id": "c",
													    "original_url": "https://fgfgfgfgfgfggf.ru/"
													}]
												`),
			Want: models.Want{
				ExpectedCode:        http.StatusCreated,
				ExpectedContentType: "application/json",
			},
		},
		{
			Name:     "Batch. No exists body.",
			Method:   http.MethodPost,
			URL:      ts.URL + "/api/shorten/batch",
			JSONBody: bytes.NewBufferString(``),
			Want: models.Want{
				ExpectedCode: http.StatusBadRequest,
			},
		},
		{
			Name:   "Batch. Not valid body original url.",
			Method: http.MethodPost,
			URL:    ts.URL + "/api/shorten/batch",
			JSONBody: bytes.NewBufferString(`
													[{
													    "correlation_id": "a",
													    "original_url": "https://123456.ru/"
													},
													{
													    "correlation_id": "b",
													    "original_url": "https://practicum.yandex.ru"
													},
													{
													    "correlation_id": "c",
													    "original_url": "sdaff/sde8%%%4325sa@.ru-213"
													}]
												`),
			Want: models.Want{
				ExpectedCode: http.StatusBadRequest,
			},
		},
		{
			Name:   "Batch. Bad json.",
			Method: http.MethodPost,
			URL:    ts.URL + "/api/shorten/batch",
			JSONBody: bytes.NewBufferString(`
													[{
													    "correlation_id": "a",
													    "original_url": "https://123456.ru/"
													},
													{
													    "correlation_id": "b",
													    "original_url": https://practicum.yandex.ru
													}]
												`),
			Want: models.Want{
				ExpectedCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			req, errReq := http.NewRequest(tt.Method, tt.URL, tt.JSONBody)
			require.NoError(t, errReq)

			client := ts.Client()
			resp, errResp := client.Do(req)
			require.NoError(t, errResp)

			defer resp.Body.Close()

			assert.Equal(t, tt.Want.ExpectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

			if tt.Want.ExpectedCode != http.StatusBadRequest {
				assert.Equal(t, tt.Want.ExpectedContentType, resp.Header.Get("content-type"), "content-type не совпадает с ожидаемым")
			}

			t.Log("=============================================================>")
		})
	}
}
