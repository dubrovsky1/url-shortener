package user

import (
	"bytes"
	"context"
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
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDeleteURL(t *testing.T) {
	logger.Initialize()

	tests := []models.TestCase{
		{
			Name: "Delete. Success.",
			Ms: models.MockStorage{
				Ctrl: gomock.NewController(t),
				DeletedURLS: []models.DeletedURLS{
					{
						ShortURL: "MlFSA8",
					},
					{
						ShortURL: "BUuk89",
					},
				},
				Error: nil,
			},
			Rp: models.RequestParams{
				Method:   http.MethodDelete,
				JSONBody: bytes.NewBufferString(`["MlFSA8","BUuk89"]`),
			},
			Want: models.Want{
				ExpectedCode: http.StatusAccepted,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			//хранилище-заглушка
			defer tt.Ms.Ctrl.Finish()

			storage := mocks.NewMockStorager(tt.Ms.Ctrl)
			serv := service.New(storage, 2, 1*time.Second)
			serv.Run(context.Background())

			//маршрутизация запроса
			r := chi.NewRouter()
			r.Delete("/api/user/urls", auth.Auth(logger.WithLogging(gzip.GzipMiddleware(DeleteURL(serv)))))

			//создание http сервера
			ts := httptest.NewServer(r)
			defer ts.Close()

			URL := ts.URL + "/api/user/urls"

			req, errReq := http.NewRequest(tt.Rp.Method, URL, tt.Rp.JSONBody)
			require.NoError(t, errReq)

			client := ts.Client()

			//добавляем куку к запросу
			jar, err := cookiejar.New(nil)
			require.NoError(t, err)

			client.Jar = jar

			tokenString, errToken := auth.BuildJWTString()
			require.NoError(t, errToken)

			c := http.Cookie{
				Name:  "userid",
				Value: tokenString,
			}
			req.AddCookie(&c)

			userID, errGetUserID := auth.GetUserID(tokenString)
			require.NoError(t, errGetUserID)

			for i := range tt.Ms.DeletedURLS {
				tt.Ms.DeletedURLS[i].UserID = userID
			}

			storage.EXPECT().DeleteURL(gomock.Any(), tt.Ms.DeletedURLS).Return(tt.Ms.Error).AnyTimes()

			resp, errResp := client.Do(req)
			require.NoError(t, errResp)

			defer resp.Body.Close()

			serv.Close()

			assert.Equal(t, tt.Want.ExpectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

			t.Log("=============================================================>")
		})
	}
}
