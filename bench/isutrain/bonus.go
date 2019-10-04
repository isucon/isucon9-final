package isutrain

import (
	"context"

	"github.com/chibiegg/isucon9-final/bench/internal/endpoint"
)

func addBonusNeighborSeats(ctx context.Context, client *Client, reservationID int) error {
	// 予約詳細を取得し、座席を列挙
	reservation, err := client.ShowReservation(ctx, reservationID)
	if err != nil {
		return err
	}

	// 座席のボーナス判定
	var (
		seats = reservation.Seats
		score = seats.GetNeighborSeatsBonus()
	)
	endpoint.AddExtraScore(endpoint.Reserve, int64(score))

	return nil
}
