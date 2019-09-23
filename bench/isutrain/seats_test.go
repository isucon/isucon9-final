package isutrain

import (
	"log"
	"testing"
)

func TestAmbigiousSearchCheck(t *testing.T) {
	tests := []struct {
		seats          TrainSeats
		wantMultiplier float64
	}{
		{},
	}

	for _, tt := range tests {
		multiplier := CalcNeighborSeatsBonus(tt.seats)
		log.Println(multiplier)
	}
}
