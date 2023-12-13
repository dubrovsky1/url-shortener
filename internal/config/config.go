package config

import "flag"

type Config struct {
	Host           string
	ResultShortURL string
}

func ParseFlags() Config {
	host := flag.String("a", "localhost:8080", "address and port to run server")
	r := flag.String("b", "http://localhost:8080/", "base address result url")

	flag.Parse()

	return Config{
		Host:           *host,
		ResultShortURL: *r,
	}
}
