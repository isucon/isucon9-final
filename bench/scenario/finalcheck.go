package scenario

import (
	"context"
	"errors"

	"github.com/chibiegg/isucon9-final/bench/payment"
	"golang.org/x/sync/errgroup"
)

var (
	ErrListReservation = errors.New("予約一覧の取得に失敗しました")
)

// FinalCheck は、課金サービスとwebappとで決済情報を突き合わせ、売上を計上します
func FinalCheck(ctx context.Context, client *payment.Client) (score int64, err error) {
	finalcheckGrp, _ := errgroup.WithContext(ctx)

	// 予約一覧と突き合わせて、不足チェック
	finalcheckGrp.Go(func() error {
		finalcheckReservations(ctx)
		return nil
	})

	// 空席であるべき座席のチェック
	finalcheckGrp.Go(func() error {
		finalcheckVocantSeats(ctx)
		return nil
	})

	// 課金が正しく反映されて入るか、予約と照らし合わせて不足チェック
	finalcheckGrp.Go(func() error {
		finalcheckPayment(ctx)
		return nil
	})

	return
}

func finalcheckReservations(ctx context.Context) error {
	// 予約キャッシュからコミット済みの予約一覧を割り出す
	// 予約確認画面を叩き、予約を突き合わせる
	return nil
}

func finalcheckVocantSeats(ctx context.Context) error {
	// 予約キャッシュから空席を割り出す
	// 座席検索をし、突き合わせる
	return nil
}

func finalcheckPayment(ctx context.Context) error {
	// 予約キャッシュから
	return nil
}
