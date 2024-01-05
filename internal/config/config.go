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
	h := flag.String("a", "localhost:8080", "address and port to run server")
	r := flag.String("b", "http://localhost:8080/", "base address result url")

	flag.Parse()

	runAddr := *h
	if sa := os.Getenv("SERVER_ADDRESS"); sa != "" {
		runAddr = sa
	}

	baseURL := *r
	if bu := os.Getenv("BASE_URL"); bu != "" {
		baseURL = bu
	}

	return Config{
		Host:           runAddr,
		ResultShortURL: baseURL,
	}
}
