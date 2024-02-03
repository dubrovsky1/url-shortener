package saveurl

import (
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
	"strings"
	"testing"
)

func TestSaveURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//определяем хранилище-заглушку
	storage := mocks.NewMockURLSaver(ctrl)

	//Передать в функцию можно что угодно - она должна это сохранить. Некорректрые url будут отсечены до сохранения в базу
	storage.EXPECT().SaveURL(gomock.Any(), "https://practicum.yandex.ru/").Return("jB9Wbk", nil).AnyTimes()
	storage.EXPECT().SaveURL(gomock.Any(), "https://yandex.ru/").Return("2Yy05g", repository.ErrUniqueIndex).AnyTimes()

	//создаем тестовый сервер, который будет проверять запросы, получаемые функцией-обработчиком хендлера saveurl
	logger.Initialize()
	r := chi.NewRouter()
	r.Post("/", logger.WithLogging(gzip.GzipMiddleware(SaveURL(storage, "http://localhost:8080/"))))
	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []models.RequestParams{
		{
			Name:   "Save url. Success.",
			Method: http.MethodPost,
			URL:    ts.URL + "/",
			Body:   "https://practicum.yandex.ru/",
			Want: models.Want{
				ExpectedCode:        http.StatusCreated,
				ExpectedContentType: "text/plain",
				ExpectedShortURL:    "jB9Wbk",
			},
		},
		{
			Name:   "Save url. Unique URL conflict.",
			Method: http.MethodPost,
			URL:    ts.URL + "/",
			Body:   "https://yandex.ru/",
			Want: models.Want{
				ExpectedCode:        http.StatusConflict,
				ExpectedContentType: "text/plain",
				ExpectedShortURL:    "2Yy05g",
			},
		},
		{
			Name:   "Save url. No exists body.",
			Method: http.MethodPost,
			URL:    ts.URL + "/",
			Body:   "",
			Want: models.Want{
				ExpectedCode: http.StatusBadRequest,
			},
		},
		{
			Name:   "Save url. Not valid body original url.",
			Method: http.MethodPost,
			URL:    ts.URL + "/",
			Body:   "sdaff/sde8%%%4325sa@.ru-213",
			Want: models.Want{
				ExpectedCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			req, errReq := http.NewRequest(tt.Method, tt.URL, strings.NewReader(tt.Body))
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
