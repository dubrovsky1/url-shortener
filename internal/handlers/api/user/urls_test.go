package user

import (
	"errors"
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
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListByUserID(t *testing.T) {
	logger.Initialize()

	tests := []models.TestCase{
		{
			Name: "Get list. Success.",
			Ms: models.MockStorage{
				Ctrl: gomock.NewController(t),
				List: []models.ShortenURL{
					{
						ShortURL:    "jB9Wbk",
						OriginalURL: "https://practicum.yandex.ru/",
					},
					{
						ShortURL:    "wqev4E",
						OriginalURL: "https://yandex.ru/",
					},
				},
				Error: nil,
			},
			Rp: models.RequestParams{
				Method: http.MethodGet,
			},
			Want: models.Want{
				ExpectedCode:        http.StatusOK,
				ExpectedContentType: "application/json",
				ExpectedJSONBody:    `[{"short_url":"jB9Wbk","original_url":"https://practicum.yandex.ru/"},{"short_url":"wqev4E","original_url":"https://yandex.ru/"}]`,
			},
		},
		{
			Name: "Get list. No content.",
			Ms: models.MockStorage{
				Ctrl:  gomock.NewController(t),
				List:  []models.ShortenURL{},
				Error: nil,
			},
			Rp: models.RequestParams{
				Method: http.MethodGet,
			},
			Want: models.Want{
				ExpectedCode: http.StatusNoContent,
			},
		},
		{
			Name: "Get list. No Cookie.",
			Ms: models.MockStorage{
				Ctrl:  gomock.NewController(t),
				List:  []models.ShortenURL{},
				Error: nil,
			},
			Rp: models.RequestParams{
				Method: http.MethodGet,
			},
			Want: models.Want{
				ExpectedCode: http.StatusUnauthorized,
			},
		},
		{
			Name: "Get list. No UserID.",
			Ms: models.MockStorage{
				Ctrl:  gomock.NewController(t),
				List:  []models.ShortenURL{},
				Error: nil,
			},
			Rp: models.RequestParams{
				Method: http.MethodGet,
			},
			Want: models.Want{
				ExpectedCode: http.StatusUnauthorized,
			},
		},
		{
			Name: "Get list. Error.",
			Ms: models.MockStorage{
				Ctrl:  gomock.NewController(t),
				List:  []models.ShortenURL{},
				Error: errors.New("error"),
			},
			Rp: models.RequestParams{
				Method: http.MethodGet,
			},
			Want: models.Want{
				ExpectedCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			//хранилище-заглушка
			defer tt.Ms.Ctrl.Finish()

			storage := mocks.NewMockStorager(tt.Ms.Ctrl)
			serv := service.New(storage)

			storage.EXPECT().ListByUserID(gomock.Any(), gomock.Any(), gomock.Any()).Return(tt.Ms.List, tt.Ms.Error).AnyTimes()

			//маршрутизация запроса
			r := chi.NewRouter()
			r.Get("/api/user/urls", auth.Auth(logger.WithLogging(gzip.GzipMiddleware(ListByUserID(serv)))))

			//создание http сервера
			ts := httptest.NewServer(r)
			defer ts.Close()

			URL := ts.URL + "/api/user/urls"

			req, errReq := http.NewRequest(tt.Rp.Method, URL, strings.NewReader(tt.Rp.Body))
			require.NoError(t, errReq)

			client := ts.Client()

			//запрет редиректа
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}

			//добавляем куку к запросу
			if tt.Name != "Get list. No Cookie." {
				jar, err := cookiejar.New(nil)
				require.NoError(t, err)

				client.Jar = jar

				if tt.Name == "Get list. No UserID." {
					c := http.Cookie{
						Name:  "userid",
						Value: "",
					}
					req.AddCookie(&c)
				} else {
					tokenString, errToken := auth.BuildJWTString()
					require.NoError(t, errToken)

					c := http.Cookie{
						Name:  "userid",
						Value: tokenString,
					}
					req.AddCookie(&c)
				}
			}

			resp, errResp := client.Do(req)
			require.NoError(t, errResp)

			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.Want.ExpectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

			if tt.Want.ExpectedCode != http.StatusBadRequest && tt.Want.ExpectedCode != http.StatusNoContent && tt.Want.ExpectedCode != http.StatusUnauthorized {
				assert.Equal(t, tt.Want.ExpectedContentType, resp.Header.Get("content-type"), "content-type не совпадает с ожидаемым")
				assert.Equal(t, tt.Want.ExpectedJSONBody, string(respBody), "Body не совпадает с ожидаемым")
			}

			t.Log("=============================================================>")
		})
	}
}
