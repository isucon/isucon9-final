package isutrain

import (
	"log"
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

func TestAmbigiousSearchCheck(t *testing.T) {
	tests := []struct {
		seats          TrainSeats
		wantMultiplier float64
	}{
		{
			seats: TrainSeats{
				&TrainSeat{
					Row:    1,
					Column: "A",
				},
				&TrainSeat{
					Row:    2,
					Column: "B",
				},
				&TrainSeat{
					Row:    3,
					Column: "C",
				},
				&TrainSeat{
					Row:    4,
					Column: "D",
				},
				&TrainSeat{
					Row:    5,
					Column: "E",
				},
			},
			wantMultiplier: 0.0,
		},
		{
			seats: TrainSeats{
				&TrainSeat{
					Row:    1,
					Column: "A",
				},
				&TrainSeat{
					Row:    1,
					Column: "B",
				},
				&TrainSeat{
					Row:    2,
					Column: "C",
				},
				&TrainSeat{
					Row:    3,
					Column: "A",
				},
				&TrainSeat{
					Row:    3,
					Column: "B",
				},
			},
			wantMultiplier: 0.2,
		},
		{
			seats: TrainSeats{
				&TrainSeat{
					Row:    1,
					Column: "A",
				},
				&TrainSeat{
					Row:    1,
					Column: "B",
				},
				&TrainSeat{
					Row:    1,
					Column: "C",
				},
				&TrainSeat{
					Row:    3,
					Column: "A",
				},
				&TrainSeat{
					Row:    3,
					Column: "B",
				},
			},
			wantMultiplier: 0.4,
		},
		{
			seats: TrainSeats{
				&TrainSeat{
					Row:    1,
					Column: "A",
				},
				&TrainSeat{
					Row:    1,
					Column: "B",
				},
				&TrainSeat{
					Row:    1,
					Column: "C",
				},
				&TrainSeat{
					Row:    1,
					Column: "D",
				},
				&TrainSeat{
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
