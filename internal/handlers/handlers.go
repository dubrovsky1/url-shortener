package handlers

import (
	"github.com/gorilla/mux"
	"io"
	"math/rand"
	"net/http"
)

var urls = make(map[string]string)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func getShortUrl() string {
	s := make([]byte, 6)
	for i := range s {
		s[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(s)
}

func GetHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
		return
	}

	shortUrl := mux.Vars(req)["id"]

	if _, ok := urls[shortUrl]; !ok {
		http.Error(res, "The short url is missing", http.StatusBadRequest)
	}

	res.Header().Set("content-type", "text/plain")
	res.Header().Set("Location", urls[shortUrl])
	res.WriteHeader(http.StatusTemporaryRedirect)
	io.WriteString(res, "")
}

func PostHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(res, "The request body is missing", http.StatusBadRequest)
	}

	shortUrl := getShortUrl()
	urls[shortUrl] = string(body)

	res.Header().Set("content-type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	io.WriteString(res, req.Host+"/"+shortUrl)
}
