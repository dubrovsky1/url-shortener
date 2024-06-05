package config

import (
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"testing"
)

func TestParseFlags(t *testing.T) {
	logger.Initialize()
	go func() {
		a := ParseFlags()
		t.Log(a)
	}()
}
