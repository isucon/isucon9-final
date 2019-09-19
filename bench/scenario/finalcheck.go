package scenario

import (
	"context"
	"errors"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/payment"
)

var (
	ErrListReservation = errors.New("予約一覧の取得に失敗しました")
)

// FinalCheck は、課金サービスとwebappとで決済情報を突き合わせ、売上を計上します
func FinalCheck(ctx context.Context, client *payment.Client) (score int64, err error) {
	var result *payment.PaymentResult
	result, err = client.Result(ctx)
	if err != nil {
		bencherror.BenchmarkErrs.AddError(bencherror.NewCriticalError(err, "課金APIから結果を取得できませんでした"))
		return
	}

	for _, rawdata := range result.RawData {
		paymentInfo := rawdata.PaymentInfo
		if !paymentInfo.IsCanceled {
			score += paymentInfo.Amount
		}
	}

	return
}
