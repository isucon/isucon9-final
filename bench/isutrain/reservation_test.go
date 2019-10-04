package isutrain

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNeighborSeatsBonus(t *testing.T) {
	tests := []struct {
		seats     ReservationSeats
		wantBonus int
	}{
		{
			seats: ReservationSeats{
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "A",
				},
				&ReservationSeat{
					SeatRow:    2,
					SeatColumn: "B",
				},
				&ReservationSeat{
					SeatRow:    3,
					SeatColumn: "C",
				},
				&ReservationSeat{
					SeatRow:    4,
					SeatColumn: "D",
				},
				&ReservationSeat{
					SeatRow:    5,
					SeatColumn: "E",
				},
			},
			wantBonus: 0,
		},
		{
			seats: ReservationSeats{
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "A",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "B",
				},
				&ReservationSeat{
					SeatRow:    2,
					SeatColumn: "C",
				},
				&ReservationSeat{
					SeatRow:    3,
					SeatColumn: "A",
				},
				&ReservationSeat{
					SeatRow:    3,
					SeatColumn: "B",
				},
			},
			wantBonus: 20,
		},
		{
			seats: ReservationSeats{
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "A",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "B",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "C",
				},
				&ReservationSeat{
					SeatRow:    3,
					SeatColumn: "A",
				},
				&ReservationSeat{
					SeatRow:    3,
					SeatColumn: "B",
				},
			},
			wantBonus: 25,
		},
		{
			seats: ReservationSeats{
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "A",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "B",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "C",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "D",
				},
				&ReservationSeat{
					SeatRow:    2,
					SeatColumn: "A",
				},
			},
			wantBonus: 20,
		},
		{
			seats: ReservationSeats{
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "A",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "B",
				},
				&ReservationSeat{
					SeatRow:    2,
					SeatColumn: "C",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "D",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "E",
				},
			},
			wantBonus: 20,
		},
		{
			seats: ReservationSeats{
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "A",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "B",
				},
				&ReservationSeat{
					SeatRow:    2,
					SeatColumn: "B",
				},
				&ReservationSeat{
					SeatRow:    2,
					SeatColumn: "C",
				},
				&ReservationSeat{
					SeatRow:    3,
					SeatColumn: "C",
				},
				&ReservationSeat{
					SeatRow:    3,
					SeatColumn: "D",
				},
			},
			wantBonus: 30,
		},
		{
			seats: ReservationSeats{
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "A",
				},
			},
			wantBonus: 0,
		},
		{
			seats: ReservationSeats{
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "A",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "B",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "C",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "D",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "E",
				},
			},
			wantBonus: 25,
		},
		{
			seats: ReservationSeats{
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "A",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "B",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "C",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "Z",
				},
				&ReservationSeat{
					SeatRow:    1,
					SeatColumn: "E",
				},
			},
			wantBonus: 0,
		},
	}

	for idx, tt := range tests {
		bonus := tt.seats.GetNeighborSeatsBonus()
		log.Printf("bonus = %d\n", bonus)
		assert.Equal(t, tt.wantBonus, bonus, "test%d", idx)
		log.Println("=====")
	}
}
