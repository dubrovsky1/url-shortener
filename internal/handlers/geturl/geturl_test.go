package geturl

import (
	"errors"
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
	"strings"
	"testing"
)

func TestGetURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//определяем хранилище-заглушку
	storage := mocks.NewMockURLGetter(ctrl)

	shortURL := "4fafrx"
	originalURL := "https://practicum.yandex.ru/"

	//Заглушка реализует определенный интерфейс, определяем, что должна возвращать функция из этого интерфейса для различных тест-кейсов
	storage.EXPECT().Get(shortURL).Return(originalURL, nil)
	storage.EXPECT().Get("aaaaaaaaaa").Return("", errors.New("the short url is missing"))

	//создаем тестовый сервер, который будет проверять запросы, получаемые функцией-обработчиком хендлера geturl
	logger.Initialize()
	r := chi.NewRouter()
	r.Get("/{id}", logger.WithLogging(gzip.GzipMiddleware(GetURL(storage))))
	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []models.RequestParams{
		{
			Name:   "Get url. Success.",
			Method: http.MethodGet,
			URL:    ts.URL + "/" + shortURL,
			Body:   "",
			Want: models.Want{
				ExpectedCode:        http.StatusTemporaryRedirect,
				ExpectedContentType: "text/plain",
				ExpectedLocation:    originalURL,
			},
		},
		{
			Name:   "Get. Not exists short url.",
			Method: http.MethodGet,
			URL:    ts.URL + "/aaaaaaaaaa",
			Body:   "",
			Want: models.Want{
				ExpectedCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			req, errReq := http.NewRequest(tt.Method, tt.URL, strings.NewReader(tt.Body))
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

			if tt.Want.ExpectedCode != http.StatusBadRequest {
				assert.Equal(t, tt.Want.ExpectedContentType, resp.Header.Get("content-type"), "content-type не совпадает с ожидаемым")
				assert.Equal(t, originalURL, resp.Header.Get("Location"), "Location не совпадает с ожидаемым")
			}

			t.Log("=============================================================>")
		})
	}
}
