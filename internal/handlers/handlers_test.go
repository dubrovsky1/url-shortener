package handlers

import (
	"github.com/dubrovsky1/url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestMainHandler(t *testing.T) {
	var shortURL string
	originalURL := "https://practicum.yandex.ru/"
	handler := Handler{Urls: *storage.New()}

	t.Run("positive test post #1", func(t *testing.T) {
		//формируем запрос
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
		w := httptest.NewRecorder()
		h := http.HandlerFunc(handler.MainHandler)
		h(w, request)

		//результат выполнения post хендлера
		result := w.Result()
		defer result.Body.Close()

		//читаем и проверяем тело ответа, достаем из него url
		body, err := io.ReadAll(result.Body)
		require.NoError(t, err)

		resultURL, err := url.Parse(string(body))
		if err != nil {
			require.Error(t, err, "Некорректный url")
		}

		shortURL = strings.TrimLeft(resultURL.Path, "/")
		log.Printf("Test post. Body: %s, shortUrl: %s\n", resultURL, shortURL)

		assert.Equal(t, http.StatusCreated, result.StatusCode, "Код ответа не совпадает с ожидаемым")
		assert.Equal(t, "text/plain", result.Header.Get("content-type"), "content-type не совпадает с ожидаемым")
	})

	log.Println("=============================================================>")

	t.Run("positive test get #1", func(t *testing.T) {
		//формируем запрос
		request := httptest.NewRequest(http.MethodGet, "/"+shortURL, nil)
		w := httptest.NewRecorder()
		h := http.HandlerFunc(handler.MainHandler)
		h(w, request)

		//результат выполнения post хендлера
		result := w.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusTemporaryRedirect, result.StatusCode, "Код ответа не совпадает с ожидаемым")
		assert.Equal(t, "text/plain", result.Header.Get("content-type"), "content-type не совпадает с ожидаемым")
		assert.Equal(t, originalURL, result.Header.Get("Location"), "Location не совпадает с ожидаемым")
	})

	log.Println("=============================================================>")

	bad_tests := []struct {
		name         string
		method       string
		target       string
		body         string
		expectedCode int
	}{
		{
			name:         "Not get and post method.",
			method:       http.MethodPut,
			target:       "/",
			body:         originalURL,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Post. No exists body.",
			method:       http.MethodPost,
			target:       "/",
			body:         "",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Post. Not valid body original url.",
			method:       http.MethodPost,
			target:       "/",
			body:         "sdaff/sde8%%%4325sa@.ru-213",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Get. Not exists short url.",
			method:       http.MethodGet,
			target:       "/aaaaaaaaaa",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range bad_tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.target, strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(handler.MainHandler)
			h(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.expectedCode, result.StatusCode, "Код ответа не совпадает с ожидаемым")
			log.Println("=============================================================>")
		})
	}

}
