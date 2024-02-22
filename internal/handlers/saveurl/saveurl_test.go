package saveurl

import (
	errs "github.com/dubrovsky1/url-shortener/internal/errors"
	"github.com/dubrovsky1/url-shortener/internal/middleware/auth"
	"github.com/dubrovsky1/url-shortener/internal/middleware/gzip"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	servmocks "github.com/dubrovsky1/url-shortener/internal/service/mocks"
	"github.com/dubrovsky1/url-shortener/internal/storage/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
				UserId:      "d110588d-369d-4e96-82eb-3a40abd234b4",
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
			serv := servmocks.NewMockStorager(tt.Ms.Ctrl)
			//serv := service.New(storage)

			//var s service.Storager

			userID, _ := uuid.Parse(tt.Ms.UserId)

			item := models.ShortenURL{
				OriginalURL: tt.Ms.OriginalURL,
				UserID:      userID,
			}

			//s := gomock.NewController(t)
			//defer s.Finish()
			//s1 := mocks2.NewMockStorager(s)
			//s1.EXPECT().SaveURL(gomock.Any(), item).Return(tt.Ms.ShortURL, tt.Ms.Error).AnyTimes()

			storage.EXPECT().SaveURL(gomock.Any(), item).Return(tt.Ms.ShortURL, tt.Ms.Error).AnyTimes()
			//s.EXPECT().SaveURL(gomock.Any(), item).Return(tt.Ms.ShortURL, tt.Ms.Error).AnyTimes()

			r := chi.NewRouter()
			r.Post("/", auth.Auth(logger.WithLogging(gzip.GzipMiddleware(SaveURL(serv, "http://localhost:8080/")))))

			ts := httptest.NewServer(r)
			defer ts.Close()

			URL := ts.URL + "/"

			req, errReq := http.NewRequest(tt.Rp.Method, URL, strings.NewReader(tt.Rp.Body))
			require.NoError(t, errReq)

			//добавляем к запросу куку
			token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDg0Nzk1NjcsIlVzZXJJRCI6ImQxMTA1ODhkLTM2OWQtNGU5Ni04MmViLTNhNDBhYmQyMzRiNCJ9.idR8V6pRw4PUjda8gQGDGCwOgWBEUlaf_oiitMq3xXM"

			c := &http.Cookie{
				Name:     "userid",
				Value:    token,
				HttpOnly: true,
				Secure:   true,
			}

			req.AddCookie(c)

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
