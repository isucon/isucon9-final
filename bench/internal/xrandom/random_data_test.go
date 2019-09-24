package xrandom

import (
	"log"
	"testing"
)

func TestRandomUseAt(t *testing.T) {
	for i := 0; i < 10; i++ {
		log.Println(GetRandomUseAt().String())
	}
}

func TestRandomSection(t *testing.T) {
	for i := 0; i < 10; i++ {
		s1, s2 := GetRandomSection()
		log.Printf("[*] s1=%s, s2=%s\n", s1, s2)
	}
}
