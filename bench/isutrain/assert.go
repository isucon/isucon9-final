package isutrain

import (
	"context"
	"net/http"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func assertListTrainSeats(resp *TrainSeatSearchResponse) error {
	if resp == nil {
		return bencherror.NewSimpleCriticalError("座席検索のレスポンスが不正です: %+v", resp)
	}

	// TODO: 席が重複して返されたら減点

	return nil
}

func assertReserve(ctx context.Context, client *Client, reserveReq *ReserveRequest, resp *ReservationResponse) error {
	if resp == nil {
		return bencherror.NewSimpleCriticalError("予約のレスポンスが不正です: %+v", resp)
	}
	if resp.ReservationID == 0 {
		return bencherror.NewSimpleApplicationError("予約のレスポンスで不正な予約IDが採番されています: %d", resp.ReservationID)
	}

	reserveGrp := &errgroup.Group{}
	// 予約一覧に正しい情報が表示されているか
	reserveGrp.Go(func() error {
		reservations, err := client.ListReservations(ctx)
		if err != nil {
			return err
		}

		for _, reservation := range reservations {
			if reservation.ReservationID == resp.ReservationID {
				// Amountのチェック
				cache, ok := ReservationCache.Reservation(reservation.ReservationID)
				if !ok {
					return bencherror.NewSimpleCriticalError("benchのキャッシュにない予約情報が存在します")
				}

				amount, err := cache.Amount()
				if err != nil {
					bencherror.SystemErrs.AddError(bencherror.NewCriticalError(err, "予約一覧画面における、予約 %dの amount取得に失敗しました", reservation.ReservationID))
					return nil
				}

				if int64(amount) != reservation.Amount {
					return bencherror.NewSimpleCriticalError("予約一覧画面における、予約 %dの amountが一致しません: want=%d, got=%d", reservation.ReservationID, amount, reservation.Amount)
				}

				return nil
			}
		}

		return bencherror.NewSimpleCriticalError("予約した内容を予約一覧画面で確認できませんでした")
	})
	// 予約確認できるか
	reserveGrp.Go(func() error {
		reservation, err := client.ShowReservation(ctx, resp.ReservationID)
		if err != nil {
			return bencherror.NewCriticalError(err, "予約した内容を予約確認画面で確認できませんでした")
		}

		// Amountのチェック
		cache, ok := ReservationCache.Reservation(reservation.ReservationID)
		if !ok {
			return bencherror.NewSimpleCriticalError("benchのキャッシュにない予約情報が存在します")
		}

		amount, err := cache.Amount()
		if err != nil {
			bencherror.SystemErrs.AddError(bencherror.NewCriticalError(err, "予約一覧画面における、予約 %dの amount取得に失敗しました", reservation.ReservationID))
			return nil
		}

		if int64(amount) != reservation.Amount {
			return bencherror.NewSimpleCriticalError("予約確認画面における、予約 %dの amountが一致しません: want=%d, got=%d", reservation.ReservationID, amount, reservation.Amount)
		}

		return nil
	})

	if err := reserveGrp.Wait(); err != nil {
		return err
	}

	return nil
}

func assertCanReserve(ctx context.Context, req *ReserveRequest, resp *ReservationResponse) error {
	lgr := zap.S()

	canReserve, err := ReservationCache.CanReserve(req)
	if err != nil {
		bencherror.SystemErrs.AddError(bencherror.NewCriticalError(err, "予約のcanReserve判定でエラー"))
		return nil
	}

	if !canReserve {
		lgr.Warnw("予約できないはず",
			"departure", req.Departure,
			"arrival", req.Arrival,
			"date", req.Date,
			"train_class", req.TrainClass,
			"train_name", req.TrainName,
			"car_num", req.CarNum,
			"seats", req.Seats,
		)
		return bencherror.NewSimpleCriticalError("予約できないはずの条件で予約が成功しました")
	}

	return nil
}

func assertCancelReservation(ctx context.Context, client *Client, reservationID int) error {
	reservations, err := client.ListReservations(ctx)
	if err != nil {
		return err
	}

	for _, reservation := range reservations {
		if reservation.ReservationID == reservationID {
			return bencherror.NewSimpleApplicationError("キャンセルされた予約が、予約一覧に列挙されています: %d", reservationID)
		}
	}

	_, err = client.ShowReservation(ctx, reservationID, StatusCodeOpt(http.StatusNotFound))
	if err != nil {
		return bencherror.NewSimpleApplicationError("キャンセルされた予約が、予約詳細で取得可能です: %d", reservationID)
	}

	return nil
}
