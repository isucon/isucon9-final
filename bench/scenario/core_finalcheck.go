package scenario

import (
	"context"
	"errors"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"

	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/payment"
	"go.uber.org/zap"
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
	lgr := zap.S()
	//
	paymentAPIResult, err := paymentClient.Result(ctx)
	if err != nil {
		return bencherror.FinalCheckErrs.AddError(bencherror.NewCriticalError(err, "決済結果を取得できませんでした. 運営に確認をお願いいたします"))
	}

	if len(paymentAPIResult.RawData) == 0 {
		return bencherror.FinalCheckErrs.AddError(bencherror.NewCriticalError(ErrInvalidReservationForPaymentAPI, "課金APIとの整合性チェックに失敗"))
	}

	eg := &errgroup.Group{}
	isutrain.ReservationCache.Range(func(reservation *isutrain.ReservationCacheEntry) {
		var (
			reservationID = reservation.ID
			amount, err   = reservation.Amount()
		)
		if err != nil {
			// FIXME: Slack通知
			lgr.Warnf("決済情報の整合性チェックでエラー: %s", err.Error())
			bencherror.FinalCheckErrs.AddError(bencherror.NewCriticalError(err, "予約の運賃取得に失敗しました"))
			return
		}

		eg.Go(func() error {
			for _, rawData := range paymentAPIResult.RawData {
				if rawData.PaymentInfo.ReservationID != reservationID {
					continue
				}
				if rawData.PaymentInfo.Amount != int64(amount) {
					lgr.Warnf("not same amount %d != %d", rawData.PaymentInfo.Amount, amount)
					return ErrInvalidReservationForPaymentAPI
				}
				return nil
			}

			// 予約IDが見つからない場合は不正
			return ErrInvalidReservationForPaymentAPI
		})
	})

	if err := eg.Wait(); err != nil {
		return bencherror.FinalCheckErrs.AddError(bencherror.NewCriticalError(err, "予約情報と決済情報で不整合を検出しました"))
	}

	return nil
}
