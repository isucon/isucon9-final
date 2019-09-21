package memory

import (
	"testing"
	"time"

	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"gopkg.in/go-playground/assert.v1"
)

func TestReservationMem(t *testing.T) {
	now := time.Now()
	mem := NewReservationMem()

	gotTests := []struct {
		req *isutrain.ReservationRequest
	}{
		{
			req: &isutrain.ReservationRequest{
				Date:       now.Add(time.Minute),
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{Row: 1, Column: "column1"},
				},
			},
		},
	}

	for _, gotTest := range gotTests {
		mem.AddFromRequest(gotTest.req)
	}

	wantTests := []struct {
		keys       []SeatMapKey
		canReserve bool
	}{
		{
			keys: []SeatMapKey{
				SeatMapKey{
					Date:       now.Add(time.Minute),
					TrainClass: "test1",
					TrainName:  "test1",
					CarNum:     1,
					Row:        1,
					Column:     "column1",
				},
			},
			canReserve: false,
		},
		{
			keys: []SeatMapKey{
				SeatMapKey{
					Date:       now.Add(time.Minute),
					TrainClass: "badtest1",
					TrainName:  "badtest1",
					CarNum:     1,
					Row:        1,
					Column:     "badc0lumn1",
				},
			},
			canReserve: true,
		},
		{
			keys: []SeatMapKey{
				SeatMapKey{
					Date:       now.Add(time.Minute),
					TrainClass: "test1",
					TrainName:  "test1",
					CarNum:     1,
					Row:        1,
					Column:     "column1",
				},
				SeatMapKey{
					Date:       now.Add(time.Minute),
					TrainClass: "badtest2",
					TrainName:  "badtest2",
					CarNum:     2,
					Row:        2,
					Column:     "badcolumn2",
				},
			},
			canReserve: false,
		},
	}

	for _, wantTest := range wantTests {
		assert.Equal(t, wantTest.canReserve, mem.CanReserve(wantTest.keys))
	}
}
