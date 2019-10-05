package isutrain

import (
	"context"
	"net/http"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// 列車検索

func assertSearchTrains(ctx context.Context, endpointPath string, resp SearchTrainsResponse) error {
	if resp == nil {
		return bencherror.NewSimpleCriticalError("GET %s: レスポンスが空です", endpointPath)
	}

	if len(resp) == 0 {
		return bencherror.NewSimpleCriticalError("GET %s: 列車が1件もヒットしませんでした", endpointPath)
	}

	for _, train := range resp {
		if ok := IsValidTrainClass(train.Class); !ok {
			return bencherror.NewSimpleCriticalError("GET %s: 列車種別が不正です: %s", endpointPath, train.Class)
		}
		if ok := IsValidStation(train.Start); !ok {
			return bencherror.NewSimpleCriticalError("GET %s: 不正な駅です: %s", endpointPath, train.Start)
		}
		if ok := IsValidStation(train.Last); !ok {
			return bencherror.NewSimpleCriticalError("GET %s: 不正な駅です: %s", endpointPath, train.Last)
		}
		if ok := IsValidStation(train.Departure); !ok {
			return bencherror.NewSimpleCriticalError("GET %s: 不正な駅です: %s", endpointPath, train.Departure)
		}
		if ok := IsValidStation(train.Arrival); !ok {
			return bencherror.NewSimpleCriticalError("GET %s: 不正な駅です: %s", endpointPath, train.Arrival)
		}
	}

	return nil
}

// 座席検索

func assertSearchTrainSeats(ctx context.Context, endpointPath string, resp *SearchTrainSeatsResponse) error {
	if resp == nil {
		return bencherror.NewSimpleCriticalError("GET %s: レスポンスが空です", endpointPath)
	}
	if _, err := time.Parse("2006/01/02", resp.Date); err != nil {
		return bencherror.NewSimpleCriticalError("GET %s: Dateの形式が不正です: %s", endpointPath, resp.Date)
	}
	if ok := IsValidTrainClass(resp.TrainClass); !ok {
		return bencherror.NewSimpleCriticalError("GET %s: 列車種別が不正です: %s", endpointPath, resp.TrainClass)
	}
	if ok := IsValidCarNumber(resp.CarNumber); !ok {
		return bencherror.NewSimpleCriticalError("GET %s: 車両番号が不正です: %d", endpointPath, resp.CarNumber)
	}

	// 席が重複して返されたら失格
	dedup := map[TrainSeat]struct{}{}
	for _, seat := range resp.Seats {
		if _, ok := dedup[*seat]; ok {
			return bencherror.NewSimpleCriticalError("GET %s: 同じ座席がレスポンスに含まれています", endpointPath)
		}
		dedup[*seat] = struct{}{}
	}

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
		return bencherror.NewSimpleCriticalError("POST %s: 予約 %dの amountが不正です: trainClass=%s,seatClass=%s,date=%s: want=%d, got=%d", endpointPath, resp.ReservationID, req.TrainClass, req.SeatClass, req.Date, amount, resp.Amount)
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
					return bencherror.NewSimpleCriticalError("POST %s: 予約 %dの amountが不正です: trainClass=%s,seatClass=%s,date=%s: want=%d, got=%d", endpointPath, reservation.ReservationID, req.TrainClass, req.SeatClass, req.Date, amount, reservation.Amount)
				}

				return nil
			}
		}

		return bencherror.NewSimpleCriticalError("POST %s: 予約した内容を予約一覧画面で確認できませんでした", endpointPath)
	})
	// 予約確認できるか
	reserveGrp.Go(func() error {
		reservation, err := client.ShowReservation(ctx, resp.ReservationID)
		if err != nil {
			return bencherror.NewCriticalError(err, "POST %s: 予約した内容を予約確認画面で確認できませんでした", endpointPath)
		}

		if len(reservation.Seats) != cache.SeatCount() {
			return bencherror.NewSimpleCriticalError("POST %s: 予約 %dの 座席が期待数確保できていません: want=%d, got=%d", cache.SeatCount(), len(reservation.Seats))
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
	if resp == nil {
		return bencherror.NewSimpleCriticalError("POST %s: レスポンスが空です", endpointPath)
	}

	return nil
}

// 予約キャンセル

func assertCancelReservation(ctx context.Context, endpointPath string, client *Client, reservationID int, resp *CancelReservationResponse) error {
	if resp == nil {
		return bencherror.NewSimpleCriticalError("POST %s: レスポンスが空です", endpointPath)
	}
	reservations, err := client.ListReservations(ctx)
	if err != nil {
		return err
	}

	for _, reservation := range reservations {
		if reservation.ReservationID == reservationID {
			return bencherror.NewSimpleCriticalError("POST %s: キャンセルされた予約が、予約一覧に列挙されています: %d", endpointPath, reservationID)
		}
	}

	_, err = client.ShowReservation(ctx, reservationID, StatusCodeOpt(http.StatusNotFound))
	if err != nil {
		return bencherror.NewSimpleCriticalError("POST %s: キャンセルされた予約が取得可能です ReservationID=%d", endpointPath, reservationID)
	}

	return nil
}
