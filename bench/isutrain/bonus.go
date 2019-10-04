package isutrain

import (
	"context"
)

func bonusNeighborSeats(ctx context.Context, client *Client, reservationID int) error {
	// 予約詳細を取得し、座席を列挙
	_, err := client.ShowReservation(ctx, reservationID)
	if err != nil {
		return err
	}

	// 座席のボーナス判定
	// seats := resp.Seats
	// score := seats.GetNeighborSeatsMultiplier()
	// endpoint.AddExtraScore(endpoint.Reserve, score)

	return nil
}
