package scenario

import (
	"context"
	"net/http"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/cache"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"golang.org/x/sync/errgroup"
)

func assertListTrainSeats(resp *isutrain.TrainSeatSearchResponse, count int) (isutrain.TrainSeats, error) {
	validSeats := isutrain.TrainSeats{}

	if resp == nil {
		return validSeats, bencherror.NewSimpleCriticalError("座席検索のレスポンスが不正です: %+v", resp)
	}

	// TODO: 席が重複して返されたら減点

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

func assertReserve(ctx context.Context, client *isutrain.Client, user *isutrain.User, reserveReq *isutrain.ReserveRequest, resp *isutrain.ReservationResponse) error {
	if resp == nil {
		return bencherror.NewSimpleCriticalError("予約のレスポンスが不正です: %+v", resp)
	}
	if resp.ReservationID == 0 {
		return bencherror.NewSimpleApplicationError("予約のレスポンスで不正な予約IDが採番されています: %d", resp.ReservationID)
	}

	reserveGrp := &errgroup.Group{}
	// 予約一覧に正しい情報が表示されているか
	reserveGrp.Go(func() error {
		reservations, err := client.ListReservations(ctx, nil)
		if err != nil {
			return err
		}

		for _, reservation := range reservations {
			if reservation.ReservationID == resp.ReservationID {
				return nil
			}
		}

		return bencherror.NewSimpleCriticalError("予約した内容を予約一覧画面で確認できませんでした")
	})
	// 予約確認できるか
	// reserveGrp.Go(func() error {
	// 	_, err := client.ShowReservation(ctx, resp.ReservationID, nil)
	// 	if err != nil {
	// 		return bencherror.NewCriticalError(err, "予約した内容を予約確認画面で確認できませんでした")
	// 	}

	// 	return nil
	// })
	if err := reserveGrp.Wait(); err != nil {
		return err
	}

	cache.ReservationCache.Add(user, reserveReq, resp.ReservationID)

	return nil
}

func assertCancelReservation(ctx context.Context, client *isutrain.Client, reservationID int) error {
	reservations, err := client.ListReservations(ctx, nil)
	if err != nil {
		return err
	}

	for _, reservation := range reservations {
		if reservation.ReservationID == reservationID {
			return bencherror.NewSimpleApplicationError("キャンセルされた予約が、予約一覧に列挙されています: %d", reservationID)
		}
	}

	_, err = client.ShowReservation(ctx, reservationID, &isutrain.ClientOption{
		WantStatusCode: http.StatusNotFound,
	})
	if err != nil {
		return bencherror.NewSimpleApplicationError("キャンセルされた予約が、予約詳細で取得可能です: %d", reservationID)
	}

	if err := cache.ReservationCache.Cancel(reservationID); err != nil {
		// FIXME: こういうベンチマーカーの異常は、利用者向けには一般的なメッセージで運営に連絡して欲しいと書き、運営向けにSlackに通知する
		return bencherror.NewCriticalError(err, "ベンチマーカーでキャッシュ不具合が発生しました. 運営に御確認お願い致します")
	}

	return nil
}
