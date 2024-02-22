package config

import (
	"flag"
	"os"
)

type Config struct {
	Host             string
	ResultShortURL   string
	FileStoragePath  string
	ConnectionString string
}

func ParseFlags() Config {
	a := flag.String("a", "localhost:8080", "address and port to run server")
	b := flag.String("b", "http://localhost:8080/", "base address result url")
	f := flag.String("f", "/tmp/short-url-db.json", "short url file")
	d := flag.String("d", "host=localhost port=5432 user=sa password=admin dbname=urls sslmode=disable", "database connection string")

	flag.Parse()

	runAddr := *a
	if sa := os.Getenv("SERVER_ADDRESS"); sa != "" {
		runAddr = sa
	}

	baseURL := *b
	if bu := os.Getenv("BASE_URL"); bu != "" {
		baseURL = bu
	}

	fileName := *f
	if fl := os.Getenv("FILE_STORAGE_PATH"); fl != "" {
		fileName = fl
	}

	connString := *d
	if cn := os.Getenv("DATABASE_DSN"); cn != "" {
		connString = cn
	}

	return Config{
		Host:             runAddr,
		ResultShortURL:   baseURL,
		FileStoragePath:  fileName,
		ConnectionString: connString,
	}
}
