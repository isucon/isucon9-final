package cache

import (
	"log"
	"testing"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/xrandom"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/stretchr/testify/assert"
)

func TestReservationMem_CanReserve_Kudari(t *testing.T) {
	now := time.Now()
	mem := newReservationCache()

	gotTests := []struct {
		reservationID int
		req           *isutrain.ReserveRequest
	}{
		{
			reservationID: 10,
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "古岡",
				Arrival:    "荒川",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "column1",
					},
				},
			},
		},
	}

	for _, gotTest := range gotTests {
		user, err := xrandom.GetRandomUser()
		assert.NoError(t, err)

		mem.Add(user, gotTest.req, gotTest.reservationID)
	}

	wantTests := []struct {
		req        *isutrain.ReserveRequest
		canReserve bool
		err        error
	}{
		// 日付
		{
			req: &isutrain.ReserveRequest{
				Date:       now.Add(2 * time.Minute),
				Departure:  "古岡",
				Arrival:    "荒川",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "column1",
					},
				},
			},
			canReserve: true,
			err:        nil,
		},
		{
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "古岡",
				Arrival:    "荒川",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "column1",
					},
				},
			},
			canReserve: false,
			err:        nil,
		},
		// 座席
		{
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "古岡",
				Arrival:    "荒川",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    9999,
						Column: "column9999",
					},
				},
			},
			canReserve: true,
			err:        nil,
		},
		{
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "古岡",
				Arrival:    "荒川",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "column1",
					},
				},
			},
			canReserve: false,
			err:        nil,
		},
		// 区間
		{
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "東京",
				Arrival:    "磯川",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "column1",
					},
				},
			},
			canReserve: false,
			err:        nil,
		},
		{
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "山田",
				Arrival:    "鳴門",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "column1",
					},
				},
			},
			canReserve: false,
			err:        nil,
		},
		{
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "東京",
				Arrival:    "山田",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "column1",
					},
				},
			},
			canReserve: false,
			err:        nil,
		},
		{
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "東京",
				Arrival:    "古岡",
				TrainClass: "badtest1",
				TrainName:  "badtest1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "badc0lumn1",
					},
				},
			},
			canReserve: true,
			err:        nil,
		},
		{ // failed
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "荒川",
				Arrival:    "鳴門",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "column1",
					},
					&isutrain.TrainSeat{
						Row:    2,
						Column: "column2",
					},
				},
			},
			canReserve: true,
			err:        nil,
		},
	}

	for idx, wantTest := range wantTests {
		log.Printf("[test%d]\n", idx)
		canReserve, err := mem.CanReserve(wantTest.req)
		assert.Equal(t, wantTest.err, err)
		assert.Equal(t, wantTest.canReserve, canReserve, "test%d failed", idx)
		log.Println("=============")
	}
}

func TestReservationMem_CanReserve_Nobori(t *testing.T) {
	now := time.Now()
	mem := newReservationCache()

	gotTests := []struct {
		reservationID int
		req           *isutrain.ReserveRequest
	}{
		{
			reservationID: 10,
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "荒川",
				Arrival:    "古岡",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "column1",
					},
				},
			},
		},
	}

	for _, gotTest := range gotTests {
		user, err := xrandom.GetRandomUser()
		assert.NoError(t, err)
		mem.Add(user, gotTest.req, gotTest.reservationID)
	}

	wantTests := []struct {
		req        *isutrain.ReserveRequest
		canReserve bool
		err        error
	}{
		// 日付
		{
			req: &isutrain.ReserveRequest{
				Date:       now.Add(2 * time.Minute),
				Departure:  "荒川",
				Arrival:    "古岡",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "column1",
					},
				},
			},
			canReserve: true,
			err:        nil,
		},
		{
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "荒川",
				Arrival:    "古岡",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "column1",
					},
				},
			},
			canReserve: false,
			err:        nil,
		},
		// 座席
		{
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "荒川",
				Arrival:    "古岡",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    9999,
						Column: "column9999",
					},
				},
			},
			canReserve: true,
			err:        nil,
		},
		{
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "荒川",
				Arrival:    "古岡",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "column1",
					},
				},
			},
			canReserve: false,
			err:        nil,
		},
		// 区間
		{
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "磯川",
				Arrival:    "東京",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "column1",
					},
				},
			},
			canReserve: false,
			err:        nil,
		},
		{
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "鳴門",
				Arrival:    "山田",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "column1",
					},
				},
			},
			canReserve: false,
			err:        nil,
		},
		{
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "山田",
				Arrival:    "東京",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "column1",
					},
				},
			},
			canReserve: false,
			err:        nil,
		},
		{
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "古岡",
				Arrival:    "東京",
				TrainClass: "badtest1",
				TrainName:  "badtest1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "badc0lumn1",
					},
				},
			},
			canReserve: true,
			err:        nil,
		},
		{ // failed
			req: &isutrain.ReserveRequest{
				Date:       now.Add(time.Minute),
				Departure:  "鳴門",
				Arrival:    "荒川",
				TrainClass: "test1",
				TrainName:  "test1",
				CarNum:     1,
				Seats: isutrain.TrainSeats{
					&isutrain.TrainSeat{
						Row:    1,
						Column: "column1",
					},
					&isutrain.TrainSeat{
						Row:    2,
						Column: "column2",
					},
				},
			},
			canReserve: true,
			err:        nil,
		},
	}

	for idx, wantTest := range wantTests {
		log.Printf("[test%d]\n", idx)
		canReserve, err := mem.CanReserve(wantTest.req)
		assert.Equal(t, wantTest.err, err)
		assert.Equal(t, wantTest.canReserve, canReserve, "test%d failed", idx)
		log.Println("=============")
	}
}
