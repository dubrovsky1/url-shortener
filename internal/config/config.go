package config

import (
	"flag"
	"os"
)

type Config struct {
	Host           string
	ResultShortURL string
}

func ParseFlags() Config {
	var baseURL string
	var runAddr string

	h := flag.String("a", "localhost:8080", "address and port to run server")
	r := flag.String("b", "http://localhost:8080/", "base address result url")

	flag.Parse()

	if sa := os.Getenv("SERVER_ADDRESS"); sa != "" {
		runAddr = sa
	} else {
		runAddr = *h
	}

	if bu := os.Getenv("BASE_URL"); bu != "" {
		baseURL = bu
	} else {
		baseURL = *r
	}

	return Config{
		Host:           runAddr,
		ResultShortURL: baseURL,
	}
}
