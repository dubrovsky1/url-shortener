package main

import (
	"github.com/dubrovsky1/url-shortener/internal/app"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
)

func main() {
	logger.Initialize()

	a := app.New()
	logger.Sugar.Infow("Flags:", "-a", a.Flags.Host, "-b", a.Flags.ResultShortURL, "-f", a.Flags.FileStoragePath, "-d", a.Flags.ConnectionString)

	a.Run()
}
