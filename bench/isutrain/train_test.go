package isutrain

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNeighborSeatsBonus(t *testing.T) {
	tests := []struct {
		seats     TrainSeats
		wantBonus int
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
			wantBonus: 0,
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
			wantBonus: 400,
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
			wantBonus: 300,
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
			wantBonus: 400,
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
					Row:    1,
					Column: "D",
				},
				&TrainSeat{
					Row:    1,
					Column: "E",
				},
			},
			wantBonus: 400,
		},
	}

	for _, tt := range tests {
		bonus := tt.seats.GetNeighborSeatsBonus()
		log.Printf("bonus = %d\n", bonus)
		assert.Equal(t, tt.wantBonus, bonus)
		log.Println("=====")
	}
}
