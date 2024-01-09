package config

import (
	"flag"
	"os"
)

type Config struct {
	Host            string
	ResultShortURL  string
	FileStoragePath string
}

func ParseFlags() Config {
	h := flag.String("a", "localhost:8080", "address and port to run server")
	r := flag.String("b", "http://localhost:8080/", "base address result url")
	f := flag.String("f", "/tmp/short-url-db.json", "short url file")

	flag.Parse()

	runAddr := *h
	if sa := os.Getenv("SERVER_ADDRESS"); sa != "" {
		runAddr = sa
	}

	baseURL := *r
	if bu := os.Getenv("BASE_URL"); bu != "" {
		baseURL = bu
	}

	fileName := *f
	if fl := os.Getenv("FILE_STORAGE_PATH"); fl != "" {
		fileName = fl
	}

	return Config{
		Host:            runAddr,
		ResultShortURL:  baseURL,
		FileStoragePath: fileName,
	}
}
