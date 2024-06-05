package storage

import (
	"github.com/dubrovsky1/url-shortener/internal/config"
	"log"
	"testing"
)

func TestGetStorage(t *testing.T) {
	flags := config.ParseFlags()
	stor, err := GetStorage(flags)
	if err != nil {
		log.Fatal("Get storage error. ", err)
	}
	stor.Close()
}
