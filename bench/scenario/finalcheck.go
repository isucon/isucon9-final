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
	ErrListReservation = errors.New("予約一覧の取得に失敗しました")
)

// FinalCheck は、課金サービスとwebappとで決済情報を突き合わせ、売上を計上します
func FinalCheck(ctx context.Context, isutrainClient *isutrain.Client, paymentClient *payment.Client) {
	finalcheckGrp := errgroup.Group{}

	// 予約一覧と突き合わせて、不足チェック
	finalcheckGrp.Go(func() error {
		finalcheckReservations(ctx, isutrainClient, paymentClient)
		return nil
	})

	// 空席であるべき座席のチェック
	// finalcheckGrp.Go(func() error {
	// 	finalcheckVocantSeats(ctx, isutrainClient)
	// 	return nil
	// })

	// 課金が正しく反映されて入るか、予約と照らし合わせて不足チェック
	finalcheckGrp.Go(func() error {
		finalcheckPayment(ctx, paymentClient)
		return nil
	})
}

func finalcheckReservations(ctx context.Context, isutrainClient *isutrain.Client, paymentClient *payment.Client) error {
	// 予約キャッシュからコミット済みの予約一覧を割り出す
	// 予約確認画面を叩き、予約を突き合わせる

	reservations, err := isutrainClient.ListReservations(ctx, nil)
	if err != nil {
		return bencherror.FinalCheckErrs.AddError(bencherror.NewApplicationError(err, "finalcheckにて予約確認画面閲覧に失敗しました"))
	}
	for _, reservation := range reservations {
		_ = reservation.ID
		// FIXME: reservationCacheにmapを持たせて、予約IDで取れるようにするべき
	}

	return nil
}

func finalcheckVocantSeats(ctx context.Context, isutrainClient *isutrain.Client) error {
	// 予約キャッシュから空席を割り出す
	// 座席検索をし、突き合わせる
	var (
		trainClass = ""
		trainName  = ""
		carNum     = -1
		departure  = ""
		arrival    = ""
	)

	// FIXME: ReservationCacheから引っ張ってこれるようにする
	isutrainClient.ListTrainSeats(
		ctx,
		trainClass,
		trainName,
		carNum,
		departure,
		arrival,
		nil,
	)
	return nil
}

func finalcheckPayment(ctx context.Context, paymentClient *payment.Client) error {
	// 予約キャッシュから

	eg := errgroup.Group{}
	cache.ReservationCache.Range(func(reservation *cache.Reservation) {
		eg.Go(func() error {
			// TODO: 課金API に予約IDで決済情報をリクエストし、amountの整合性を比較
			// 挙動がおかしい場合、失格とする
			return nil
		})
	})

	return eg.Wait()
}
