package scenario

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

func TestVagueSearchCheck(t *testing.T) {
	tests := []struct {
		seats          isutrain.TrainSeats
		wantMultiplier float64
	}{
		{
			seats: isutrain.TrainSeats{
				&isutrain.TrainSeat{
					Row:    1,
					Column: "A",
				},
				&isutrain.TrainSeat{
					Row:    2,
					Column: "B",
				},
				&isutrain.TrainSeat{
					Row:    3,
					Column: "C",
				},
				&isutrain.TrainSeat{
					Row:    4,
					Column: "D",
				},
				&isutrain.TrainSeat{
					Row:    5,
					Column: "E",
				},
			},
			wantMultiplier: 0.0,
		},
		{
			seats: isutrain.TrainSeats{
				&isutrain.TrainSeat{
					Row:    1,
					Column: "A",
				},
				&isutrain.TrainSeat{
					Row:    1,
					Column: "B",
				},
				&isutrain.TrainSeat{
					Row:    2,
					Column: "C",
				},
				&isutrain.TrainSeat{
					Row:    3,
					Column: "A",
				},
				&isutrain.TrainSeat{
					Row:    3,
					Column: "B",
				},
			},
			wantMultiplier: 0.2,
		},
		{
			seats: isutrain.TrainSeats{
				&isutrain.TrainSeat{
					Row:    1,
					Column: "A",
				},
				&isutrain.TrainSeat{
					Row:    1,
					Column: "B",
				},
				&isutrain.TrainSeat{
					Row:    1,
					Column: "C",
				},
				&isutrain.TrainSeat{
					Row:    3,
					Column: "A",
				},
				&isutrain.TrainSeat{
					Row:    3,
					Column: "B",
				},
			},
			wantMultiplier: 0.4,
		},
		{
			seats: isutrain.TrainSeats{
				&isutrain.TrainSeat{
					Row:    1,
					Column: "A",
				},
				&isutrain.TrainSeat{
					Row:    1,
					Column: "B",
				},
				&isutrain.TrainSeat{
					Row:    1,
					Column: "C",
				},
				&isutrain.TrainSeat{
					Row:    1,
					Column: "D",
				},
				&isutrain.TrainSeat{
					Row:    2,
					Column: "A",
				},
			},
			wantMultiplier: 0.9,
		},
	}

	for _, tt := range tests {
		multiplier := CalcNeighborSeatsBonus(tt.seats)
		log.Printf("multiplier = %f\n", multiplier)
		assert.Equal(t, tt.wantMultiplier, multiplier)
		log.Println("=====")
	}
}
