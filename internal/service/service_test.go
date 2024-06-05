package service

import (
	"github.com/dubrovsky1/url-shortener/internal/config"
	"github.com/dubrovsky1/url-shortener/internal/storage"
	"log"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	flags := config.ParseFlags()
	stor, err := storage.GetStorage(flags)
	if err != nil {
		log.Fatal("Get storage error. ", err)
	}
	serv := New(stor, 10, time.Second*10)
	t.Log(serv.deleteBatchSize)
}
