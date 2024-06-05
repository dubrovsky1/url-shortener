package app

import (
	"testing"
)

func BenchmarkClose(b *testing.B) {
	//logger.Initialize()
	//b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//a := New()
		//a.Close()
	}
}
