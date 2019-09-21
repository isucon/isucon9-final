package cache

import (
	"log"
	"testing"
	"time"

	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"gopkg.in/go-playground/assert.v1"
)

func TestReservationMem_Commit(t *testing.T) {
	now := time.Now()
	mem := NewReservationMem()

	gotTests := []struct {
		reservationID string
		req           *isutrain.ReservationRequest
	}{
		{
			reservationID: "",
			req: &isutrain.ReservationRequest{
				Date:        now.Add(time.Minute),
				Origin:      "古岡",
				Destination: "荒川",
				TrainClass:  "test1",
				TrainName:   "test1",
				CarNum:      1,
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
		mem.Add(gotTest.req, gotTest.reservationID)
	}

	wantTests := []struct {
		req        *isutrain.ReservationRequest
		canReserve bool
		err        error
	}{
		{
			req: &isutrain.ReservationRequest{
				Date:        now.Add(time.Minute),
				Origin:      "東京",
				Destination: "山田",
				TrainClass:  "test1",
				TrainName:   "test1",
				CarNum:      1,
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
			req: &isutrain.ReservationRequest{
				Date:        now.Add(time.Minute),
				Origin:      "気川",
				Destination: "名古屋",
				TrainClass:  "badtest1",
				TrainName:   "badtest1",
				CarNum:      1,
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
		{
			req: &isutrain.ReservationRequest{
				Date:        now.Add(time.Minute),
				Origin:      "東京",
				Destination: "古岡",
				TrainClass:  "test1",
				TrainName:   "test1",
				CarNum:      1,
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
			canReserve: false,
			err:        nil,
		},
	}

	for _, wantTest := range wantTests {
		canReserve, err := mem.CanReserve(wantTest.req)
		assert.Equal(t, wantTest.err, err)
		assert.Equal(t, wantTest.canReserve, canReserve)
		log.Println("=============")
	}
}

func TestReservationMem_Cancel(t *testing.T) {

}
