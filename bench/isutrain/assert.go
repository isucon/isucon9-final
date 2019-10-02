package isutrain

import (
	"context"
	"net/http"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// 座席検索

func assertSearchTrainSeats(ctx context.Context, endpointPath string, resp *SearchTrainSeatsResponse) error {
	if resp == nil {
		return bencherror.NewSimpleCriticalError("GET %s: レスポンスが不正です: %+v", endpointPath, resp)
	}

	// TODO: 席が重複して返されたら減点

	return nil
}

// 予約

func assertReserve(ctx context.Context, endpointPath string, client *Client, req *ReserveRequest, resp *ReserveResponse) error {
	lgr := zap.S()
	if resp == nil {
		return bencherror.NewSimpleCriticalError("POST %s: レスポンスが不正です: %+v", endpointPath, resp)
	}

	cache, ok := ReservationCache.Reservation(resp.ReservationID)
	if !ok {
		bencherror.SystemErrs.AddError(bencherror.NewSimpleCriticalError("POST %s: 予約キャッシュの取得に失敗: ReservationID=%d", endpointPath, resp.ReservationID))
		return nil
	}

	amount, err := cache.Amount()
	if err != nil {
		bencherror.SystemErrs.AddError(bencherror.NewSimpleCriticalError("POST %s: 予約のamount取得に失敗: resp.IsOK=%v, resp.ReservationID=%d, resp.Amount=%d", endpointPath, resp.IsOk, resp.ReservationID, resp.Amount))
		return nil
	}

	// レスポンスに含まれるamountは正しいか
	if resp.Amount != amount {
		lgr.Warnf("amountが不正になった!? reservationID=%d, want=%d, got=%d", resp.ReservationID, amount, resp.Amount)
		return bencherror.NewSimpleApplicationError("POST %s: amountが不正です: want=%d, got=%d", endpointPath, amount, resp.Amount)
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
				if amount != reservation.Amount {
					return bencherror.NewSimpleApplicationError("POST %s: 予約 %dの amountが不正です: want=%d, got=%d", endpointPath, reservation.ReservationID, amount, reservation.Amount)
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
			return bencherror.NewCriticalError(err, "POST %s: 予約した内容を予約確認画面で確認できませんでした", endpointPath)
		}

		if amount != reservation.Amount {
			return bencherror.NewSimpleCriticalError("POST %s: 予約 %dの amountが一致しません: want=%d, got=%d", endpointPath, reservation.ReservationID, amount, reservation.Amount)
		}

		return nil
	})

	if err := reserveGrp.Wait(); err != nil {
		return err
	}

	return nil
}

func assertCanReserve(ctx context.Context, endpointPath string, req *ReserveRequest, resp *ReserveResponse) error {
	lgr := zap.S()

	canReserve, err := ReservationCache.CanReserve(req)
	if err != nil {
		bencherror.SystemErrs.AddError(bencherror.NewCriticalError(err, "POST %s: 予約可能判定でエラーが発生しました", endpointPath))
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
		return bencherror.NewSimpleCriticalError("POST %s: 予約できないはずの条件で予約が成功しました", endpointPath)
	}

	return nil
}

// 予約コミット

func assertCommitReservation(ctx context.Context, endpointPath string, resp *CommitReservationResponse) error {
	if !resp.IsOK {
		return bencherror.NewSimpleCriticalError("POST %s: is_ok がfalseです", endpointPath)
	}

	return nil
}

// 予約キャンセル

func assertCancelReservation(ctx context.Context, endpointPath string, client *Client, reservationID int) error {
	reservations, err := client.ListReservations(ctx)
	if err != nil {
		return err
	}

	for _, reservation := range reservations {
		if reservation.ReservationID == reservationID {
			return bencherror.NewSimpleApplicationError("POST %s: キャンセルされた予約が、予約一覧に列挙されています: %d", endpointPath, reservationID)
		}
	}

	_, err = client.ShowReservation(ctx, reservationID, StatusCodeOpt(http.StatusNotFound))
	if err != nil {
		return bencherror.NewSimpleApplicationError("POST %s: キャンセルされた予約が取得可能です ReservationID=%d", endpointPath, reservationID)
	}

	return nil
}
