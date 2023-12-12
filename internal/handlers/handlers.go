package handlers

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

var urls = make(map[string]string)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func getShortURL() string {
	s := make([]byte, 6)
	for i := range s {
		s[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(s)
}

func MainHandler(res http.ResponseWriter, req *http.Request) {
	log.Printf("Request Log. Method: %s\n", req.Method)

	if req.Method == http.MethodPost {
		postHandler(res, req)
	} else if req.Method == http.MethodGet {
		getHandler(res, req)
	} else {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
	}
}

func getHandler(res http.ResponseWriter, req *http.Request) {
	shortURL := strings.TrimLeft(req.URL.Path, "/")
	log.Printf("Request Log. shortURL: %s\n", shortURL)

	if _, ok := urls[shortURL]; !ok {
		http.Error(res, "The short url is missing", http.StatusBadRequest)
		return
	}

	res.Header().Set("content-type", "text/plain")
	res.Header().Set("Location", urls[shortURL])
	res.WriteHeader(http.StatusTemporaryRedirect)

	log.Printf("Response Log. content-type: %s, Location: %s\n", res.Header().Get("content-type"), res.Header().Get("Location"))
}

func postHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	log.Printf("Request Log. Body: %s\n", body)

	if err != nil {
		http.Error(res, "The request body is missing", http.StatusBadRequest)
		return
	}

	//нужно ли проверять ссылку из тела на корректность или там может быть любой текст?

	//гененрируем короткую ссылку
	shortURL := getShortURL()

	//запоминаем url, соответствующий короткой ссылке
	urls[shortURL] = string(body)

	//формируем тело ответа и проверяем на валидность
	responseBody := "http://" + req.Host + "/" + shortURL

	if _, e := url.Parse(responseBody); e != nil {
		http.Error(res, "Not valid result URL", http.StatusBadRequest)
		return
	}

	res.Header().Set("content-type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	io.WriteString(res, responseBody)

	log.Printf("Response Log. content-type: %s, shortURL: %s\n", res.Header().Get("content-type"), shortURL)
}
