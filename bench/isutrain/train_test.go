package isutrain

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNeighborSeatsMultiplier(t *testing.T) {
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
			wantMultiplier: 1.0,
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
			wantMultiplier: 1.2,
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
			wantMultiplier: 1.4,
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
			wantMultiplier: 1.9,
		},
	}

	for _, tt := range tests {
		multiplier := tt.seats.GetNeighborSeatsMultiplier()
		log.Printf("multiplier = %f\n", multiplier)
		assert.Equal(t, tt.wantMultiplier, multiplier)
		log.Println("=====")
	}
}
