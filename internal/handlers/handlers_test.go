package handlers

import (
	"github.com/dubrovsky1/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
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

type Want struct {
	expectedCode        int
	expectedContentType string
	expectedLocation    string
	expectedBody        string
}

type RequestParams struct {
	name   string
	method string
	url    string
	body   string
	want   Want
}

var shortURL string
var originalURL = "https://practicum.yandex.ru/"
var h = Handler{Urls: *storage.New()}

func getRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.SaveURL)
	r.Get("/{id}", h.GetURL)
	return r
}

func testRequest(t *testing.T, ts *httptest.Server, method, url string, body string) (*http.Response, string) {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestSaveURL(t *testing.T) {
	log.Println("=============================================================>")

	//запускаем тестовый сервер
	r := getRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []RequestParams{
		{
			name:   "Post. Save url.",
			method: http.MethodPost,
			url:    ts.URL + "/",
			body:   originalURL,
			want: Want{
				expectedCode:        http.StatusCreated,
				expectedContentType: "text/plain",
			},
		},
		{
			name:   "Post. No exists body.",
			method: http.MethodPost,
			url:    ts.URL + "/",
			body:   "",
			want: Want{
				expectedCode: http.StatusBadRequest,
			},
		},
		{
			name:   "Post. Not valid body original url.",
			method: http.MethodPost,
			url:    ts.URL + "/",
			body:   "sdaff/sde8%%%4325sa@.ru-213",
			want: Want{
				expectedCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, respBody := testRequest(t, ts, tt.method, tt.url, tt.body)

			assert.Equal(t, tt.want.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

			if tt.want.expectedCode != http.StatusBadRequest {
				u, errParseBody := url.Parse(string(respBody))
				require.NoError(t, errParseBody)

				shortURL = strings.TrimLeft(u.Path, "/")
				log.Printf("Test Log. RespBody: %s, URL: %s, ShortURL: %s\n", string(respBody), ts.URL, shortURL)

				assert.Equal(t, tt.want.expectedContentType, resp.Header.Get("content-type"), "content-type не совпадает с ожидаемым")
				assert.Equal(t, ts.URL+"/"+shortURL, string(respBody), "Body не совпадает с ожидаемым")
			}

			log.Println("=============================================================>")
		})
	}

}

func TestGetURL(t *testing.T) {
	log.Println("=============================================================>")

	//запускаем тестовый сервер
	r := getRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []RequestParams{
		{
			name:   "Get. Get url.",
			method: http.MethodGet,
			url:    ts.URL + "/" + shortURL,
			body:   "",
			want: Want{
				expectedCode:        http.StatusTemporaryRedirect,
				expectedContentType: "text/plain",
				expectedLocation:    originalURL,
			},
		},
		{
			name:   "Get. Not exists short url.",
			method: http.MethodGet,
			url:    ts.URL + "/aaaaaaaaaa",
			body:   "",
			want: Want{
				expectedCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			log.Printf("Test Log. URL: %s\n", tt.url)

			//отправляем запросы
			resp, _ := testRequest(t, ts, tt.method, tt.url, tt.body)

			assert.Equal(t, tt.want.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

			if tt.want.expectedCode != http.StatusBadRequest {
				log.Printf("RespLocation: %s\n", resp.Header.Get("Location"))

				assert.Equal(t, tt.want.expectedContentType, resp.Header.Get("content-type"), "content-type не совпадает с ожидаемым")
				assert.Equal(t, originalURL, resp.Header.Get("Location"), "Location не совпадает с ожидаемым")
			}

			log.Println("=============================================================>")
		})
	}
}
