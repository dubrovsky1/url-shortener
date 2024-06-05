package app

import (
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"testing"
)

func TestNew(t *testing.T) {
	logger.Initialize()
	go func() {
		a := New()
		a.Close()
	}()
}
