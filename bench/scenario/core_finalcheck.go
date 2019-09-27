package scenario

import (
	"context"
	"errors"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/cache"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/payment"
	"golang.org/x/sync/errgroup"
)

var (
	ErrListReservation                 = errors.New("予約一覧の取得に失敗しました")
	ErrInvalidReservationForPaymentAPI = errors.New("課金APIと予約の整合性が取れていません")
	ErrInvalidReservationForBenchCache = errors.New("予約における計算結果が")
	ErrNoReservationPayments           = errors.New("予約に紐づく課金情報がありません")
)

// FinalCheck は、課金サービスとwebappとで決済情報を突き合わせ、売上を計上します
func FinalCheck(ctx context.Context, isutrainClient *isutrain.Client, paymentClient *payment.Client) {
	// 予約一覧と突き合わせて、不足チェック
	finalcheckPayment(ctx, paymentClient)
}

func finalcheckPayment(ctx context.Context, paymentClient *payment.Client) error {
	//
	paymentAPIResult, err := paymentClient.Result(ctx)
	if err != nil {
		return bencherror.FinalCheckErrs.AddError(bencherror.NewCriticalError(err, "決済結果を取得できませんでした. 運営に確認をお願いいたします"))
	}

	if len(paymentAPIResult.RawData) == 0 {
		return bencherror.FinalCheckErrs.AddError(bencherror.NewCriticalError(ErrInvalidReservationForPaymentAPI, "課金APIとの整合性チェックに失敗"))
	}

	eg := &errgroup.Group{}
	cache.ReservationCache.Range(func(reservation *cache.Reservation) {
		var (
			reservationID = reservation.ID
			amount, err   = reservation.Amount()
		)
		if err != nil {
			// FIXME: Slack通知
		}

		eg.Go(func() error {
			for _, rawData := range paymentAPIResult.RawData {
				if rawData.PaymentInfo.ReservationID != reservationID {
					continue
				}
				if rawData.PaymentInfo.Amount != int64(amount) {
					return ErrInvalidReservationForPaymentAPI
				}
			}
			return nil
		})
	})

	if err := eg.Wait(); err != nil {
		return bencherror.FinalCheckErrs.AddError(bencherror.NewCriticalError(err, "予約情報と決済情報で不整合を検出しました"))
	}

	return nil
}

func finalcheckVocantSeats(ctx context.Context, isutrainClient *isutrain.Client) error {
	// 予約キャッシュから空席を割り出す
	// 座席検索をし、突き合わせる

	return nil
}
