package scenario

import (
	"context"
	"errors"

	"github.com/chibiegg/isucon9-final/bench/payment"
)

var (
	ErrListReservation = errors.New("予約一覧の取得に失敗しました")
)

// FinalCheck は、課金サービスとwebappとで決済情報を突き合わせ、売上を計上します
func FinalCheck(ctx context.Context, client *payment.Client) (score int64, err error) {
	// FIXME: 予約情報の整合性チェック
	return
}
