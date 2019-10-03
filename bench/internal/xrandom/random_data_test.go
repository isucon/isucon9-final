package xrandom

import (
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestGetRandomNumberOfPeople(t *testing.T) {
	for i := 0; i < 10; i++ {
		adult, child := GetRandomNumberOfPeople()
		log.Printf("adult=%d, child=%d", adult, child)
	}
}

func TestRandomUseAt(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	assert.NoError(t, config.SetAvailReserveDays(30))
	for i := 0; i < 10; i++ {
		log.Println(GetRandomUseAt().String())
	}
}

func TestRandomUseAtByOlympicDate(t *testing.T) {
	for i := 0; i < 10; i++ {
		log.Println(GetRandomUseAtByOlympicDate().String())
	}
}

func TestRandomSection(t *testing.T) {
	for i := 0; i < 10; i++ {
		s1, s2 := GetRandomSection()
		log.Printf("[*] s1=%s, s2=%s\n", s1, s2)
	}
}

func TestRandomSectionWithTokyo(t *testing.T) {
	for i := 0; i < 10; i++ {
		s1, s2 := GetRandomSectionWithTokyo()
		log.Printf("[*] s1=%s, s2=%s\n", s1, s2)
	}
}
