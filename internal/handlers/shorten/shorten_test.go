package shorten

import (
	"bytes"
	"github.com/dubrovsky1/url-shortener/internal/handlers/mocks"
	"github.com/dubrovsky1/url-shortener/internal/middleware/gzip"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	repository "github.com/dubrovsky1/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShorten(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//определяем хранилище-заглушку
	storage := mocks.NewMockURLSaver(ctrl)

	//Передать в функцию можно что угодно - она должна это сохранить. Некорректрые url будут отсечены до сохранения в базу
	storage.EXPECT().SaveURL(gomock.Any(), "https://practicum.yandex.ru/").Return("jB9Wbk", nil).AnyTimes()
	storage.EXPECT().SaveURL(gomock.Any(), "https://yandex.ru/").Return("2Yy05g", repository.ErrUniqueIndex).AnyTimes()

	//создаем тестовый сервер, который будет проверять запросы, получаемые функцией-обработчиком хендлера shorten
	logger.Initialize()
	r := chi.NewRouter()
	r.Post("/api/shorten", logger.WithLogging(gzip.GzipMiddleware(Shorten(storage, "http://localhost:8080/"))))
	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []models.RequestParams{
		{
			Name:     "Shorten save url. Success.",
			Method:   http.MethodPost,
			URL:      ts.URL + "/api/shorten",
			JSONBody: bytes.NewBufferString(`{"url": "https://practicum.yandex.ru/"}`),
			Want: models.Want{
				ExpectedCode:        http.StatusCreated,
				ExpectedContentType: "application/json",
				ExpectedShortURL:    "jB9Wbk",
			},
		},
		{
			Name:     "Shorten save url. Unique URL conflict.",
			Method:   http.MethodPost,
			URL:      ts.URL + "/api/shorten",
			JSONBody: bytes.NewBufferString(`{"url": "https://yandex.ru/"}`),
			Want: models.Want{
				ExpectedCode:        http.StatusConflict,
				ExpectedContentType: "application/json",
				ExpectedShortURL:    "2Yy05g",
			},
		},
		{
			Name:     "Shorten save url. No exists body.",
			Method:   http.MethodPost,
			URL:      ts.URL + "/api/shorten",
			JSONBody: bytes.NewBufferString(""),
			Want: models.Want{
				ExpectedCode: http.StatusBadRequest,
			},
		},
		{
			Name:     "Shorten save url. Not valid json.",
			Method:   http.MethodPost,
			URL:      ts.URL + "/api/shorten",
			JSONBody: bytes.NewBufferString(`{"url": https://practicum.yandex.ru}`),
			Want: models.Want{
				ExpectedCode: http.StatusBadRequest,
			},
		},
		{
			Name:     "Shorten save url. Not valid body original url.",
			Method:   http.MethodPost,
			URL:      ts.URL + "/api/shorten",
			JSONBody: bytes.NewBufferString(`{"url": "sdaff/sde8%%%4325sa@.ru-213"}`),
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

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.Want.ExpectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

			if tt.Want.ExpectedCode != http.StatusBadRequest {
				assert.Equal(t, tt.Want.ExpectedContentType, resp.Header.Get("content-type"), "content-type не совпадает с ожидаемым")
				assert.Equal(t, `{"result":"`+ts.URL+`/`+tt.Want.ExpectedShortURL+`"}`, string(respBody), "Body не совпадает с ожидаемым")
			}

			t.Log("=============================================================>")
		})
	}

}
