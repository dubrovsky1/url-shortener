package generator

import (
	"testing"
)

func TestGetShortURL(t *testing.T) {
	var arr []string
	s := GetShortURL()
	arr = append(arr, s)
	t.Log(arr)
}
