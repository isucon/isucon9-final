package scenario

import (
	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

func assertListTrainSeats(resp *isutrain.TrainSeatSearchResponse, count int) (isutrain.TrainSeats, error) {
	validSeats := isutrain.TrainSeats{}

	if resp == nil {
		return validSeats, bencherror.NewSimpleCriticalError("座席検索のレスポンスが不正です: %+v", resp)
	}

	for _, seat := range resp.Seats {
		if len(validSeats) == count {
			break
		}
		if !seat.IsOccupied {
			validSeats = append(validSeats, seat)
		}
	}

	return validSeats, nil
}

func assertReserve(resp *isutrain.ReservationResponse) error {
	if resp == nil {
		return bencherror.NewSimpleCriticalError("予約のレスポンスが不正です: %+v", resp)
	}
	if resp.ReservationID == 0 {
		return bencherror.NewSimpleApplicationError("予約のレスポンスで不正な予約IDが採番されています: %d", resp.ReservationID)
	}

	// FIXME: IsOKのチェック

	return nil
}
