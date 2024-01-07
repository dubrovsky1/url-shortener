package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/dubrovsky1/url-shortener/internal/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/dubrovsky1/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"io"
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
}

type RequestParams struct {
	name     string
	method   string
	url      string
	body     string
	jsonBody *bytes.Buffer
	want     Want
}

var shortURL string
var originalURL = "https://practicum.yandex.ru/"

var h = Handler{Urls: *storage.New()}

func getRouter() chi.Router {
	logger.Initialize()
	r := chi.NewRouter()
	r.Post("/", h.SaveURL)
	r.Post("/api/shorten", logger.WithLogging(h.Shorten))
	r.Get("/{id}", h.GetURL)
	return r
}

func testRequest(t *testing.T, ts *httptest.Server, method, url string, body string, jsonBody *bytes.Buffer) *http.Response {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if jsonBody != nil {
		req, err = http.NewRequest(method, url, jsonBody)
	}
	require.NoError(t, err)

	//запрет редиректа
	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	return resp
}

func TestSaveURL(t *testing.T) {
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
			resp := testRequest(t, ts, tt.method, tt.url, tt.body, nil)
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.want.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

			if tt.want.expectedCode != http.StatusBadRequest {
				u, errParseBody := url.Parse(string(respBody))
				require.NoError(t, errParseBody)

				shortURL = strings.TrimLeft(u.Path, "/")
				t.Logf("Test Log. RespBody: %s, URL: %s, ShortURL: %s\n", respBody, ts.URL, shortURL)

				assert.Equal(t, tt.want.expectedContentType, resp.Header.Get("content-type"), "content-type не совпадает с ожидаемым")
				assert.Equal(t, ts.URL+"/"+shortURL, string(respBody), "Body не совпадает с ожидаемым")
			}

			t.Log("=============================================================>")
		})
	}

}

func TestGetURL(t *testing.T) {
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
			t.Logf("Test Log. URL: %s\n", tt.url)

			//отправляем запросы
			resp := testRequest(t, ts, tt.method, tt.url, tt.body, nil)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

			if tt.want.expectedCode != http.StatusBadRequest {
				assert.Equal(t, tt.want.expectedContentType, resp.Header.Get("content-type"), "content-type не совпадает с ожидаемым")
				assert.Equal(t, originalURL, resp.Header.Get("Location"), "Location не совпадает с ожидаемым")
			}

			t.Log("=============================================================>")
		})
	}
}

func TestShorten(t *testing.T) {
	r := getRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []RequestParams{
		{
			name:     "Post Shorten. Save url.",
			method:   http.MethodPost,
			url:      ts.URL + "/api/shorten",
			jsonBody: bytes.NewBufferString(`{"url": "https://practicum.yandex.ru"}`),
			want: Want{
				expectedCode:        http.StatusCreated,
				expectedContentType: "application/json",
			},
		},
		{
			name:     "Post Shorten. No exists body.",
			method:   http.MethodPost,
			url:      ts.URL + "/api/shorten",
			jsonBody: bytes.NewBufferString(""),
			want: Want{
				expectedCode: http.StatusBadRequest,
			},
		},
		{
			name:     "Post Shorten. Not valid json.",
			method:   http.MethodPost,
			url:      ts.URL + "/api/shorten",
			jsonBody: bytes.NewBufferString(`{"url": https://practicum.yandex.ru}`),
			want: Want{
				expectedCode: http.StatusBadRequest,
			},
		},
		{
			name:     "Post Shorten. Not valid body original url.",
			method:   http.MethodPost,
			url:      ts.URL + "/api/shorten",
			jsonBody: bytes.NewBufferString(`{"url": "sdaff/sde8%%%4325sa@.ru-213"}`),
			want: Want{
				expectedCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := testRequest(t, ts, tt.method, tt.url, "", tt.jsonBody)
			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

			if tt.want.expectedCode != http.StatusBadRequest {
				//json из тела ответа
				var r models.Response
				err = json.Unmarshal(respBody, &r)
				require.NoError(t, err)

				//сформированный url из ответа
				u, errParseBody := url.Parse(r.Result)
				require.NoError(t, errParseBody)

				shortURL = strings.TrimLeft(u.Path, "/")
				t.Logf("Test Log. RespBody: %s, URL: %s, ShortURL: %s, ExpectedBody: %s\n", respBody, ts.URL, shortURL, `{"result":"`+ts.URL+`/`+shortURL+`"}`)

				assert.Equal(t, tt.want.expectedContentType, resp.Header.Get("content-type"), "content-type не совпадает с ожидаемым")
				assert.Equal(t, `{"result":"`+ts.URL+`/`+shortURL+`"}`, string(respBody), "Body не совпадает с ожидаемым")
			}

			t.Log("=============================================================>")
		})
	}

}
