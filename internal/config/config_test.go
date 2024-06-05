package config

import (
	"testing"
)

func BenchmarkParseFlags(b *testing.B) {
	//var flags []Config
	//flags := ParseFlags()
	for i := 0; i < b.N; i++ {
		//	flag := ParseFlags()
		//	flags = append(flags, flag)
	}
	//b.Log(flags)
}
