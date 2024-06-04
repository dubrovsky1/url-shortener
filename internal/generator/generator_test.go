package generator

import (
	"testing"
)

func BenchmarkGetShortURL(b *testing.B) {
	var arr []string
	for i := 0; i < b.N; i++ {
		s := GetShortURL()
		arr = append(arr, s)
	}
	b.Log(arr)
}
