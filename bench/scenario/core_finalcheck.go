package scenario

import (
	"context"
	"errors"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/cache"
	"github.com/chibiegg/isucon9-final/bench/internal/xrandom"
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
	return // 未実装
	// finalcheckGrp := errgroup.Group{}

	// // 予約一覧と突き合わせて、不足チェック
	// finalcheckGrp.Go(func() error {
	// 	finalcheckReservations(ctx, isutrainClient, paymentClient)
	// 	return nil
	// })

	// // 空席であるべき座席のチェック
	// // finalcheckGrp.Go(func() error {
	// // 	finalcheckVocantSeats(ctx, isutrainClient)
	// // 	return nil
	// // })
	// lgr := zap.S()
	// if err := finalcheckGrp.Wait(); err != nil {
	// 	lgr.Warnf("最終チェックでエラーが発生しました: %+v", err)
	// }
}

func finalcheckReservations(ctx context.Context, isutrainClient *isutrain.Client, paymentClient *payment.Client) error {
	// 予約キャッシュからコミット済みの予約一覧を割り出す
	// 予約確認画面を叩き、予約を突き合わせる

	//
	paymentAPIResult, err := paymentClient.Result(ctx)
	if err != nil {
		return bencherror.FinalCheckErrs.AddError(bencherror.NewCriticalError(err, "決済結果を取得できませんでした. 運営に確認をお願いいたします"))
	}

	user, err := xrandom.GetRandomUser()
	if err != nil {
		return bencherror.FinalCheckErrs.AddError(bencherror.NewCriticalError(err, "ユーザを作成できませんでした. 運営に確認をお願いいたします"))
	}
	if err := isutrainClient.Signup(ctx, user.Email, user.Password, nil); err != nil {
		return bencherror.FinalCheckErrs.AddError(bencherror.NewApplicationError(err, "登録できませんでした"))
	}

	if err := isutrainClient.Login(ctx, user.Email, user.Password, nil); err != nil {
		return bencherror.FinalCheckErrs.AddError(bencherror.NewApplicationError(err, "ログインできませんでした"))
	}

	reservations, err := isutrainClient.ListReservations(ctx, nil)
	if err != nil {
		return bencherror.FinalCheckErrs.AddError(bencherror.NewApplicationError(err, "finalcheckにて予約確認画面閲覧に失敗しました"))
	}

	if len(paymentAPIResult.RawData) == 0 {
		return bencherror.FinalCheckErrs.AddError(bencherror.NewCriticalError(ErrInvalidReservationForPaymentAPI, "課金APIとの整合性チェックに失敗"))
	}

	eg := &errgroup.Group{}
	for _, reservation := range reservations {
		var (
			reservationID = reservation.ReservationID
			amount        = reservation.Amount
		)
		// 課金APIと突き合わせる
		eg.Go(func() error {
			for _, rawData := range paymentAPIResult.RawData {
				if rawData.PaymentInfo.ReservationID != reservationID {
					continue
				}
				if rawData.PaymentInfo.Amount != amount {
					return ErrInvalidReservationForPaymentAPI
				}
			}
			return nil
		})
		// ベンチマーカーのキャッシュと突き合わせる
		// eg.Go(func() error {

		// })
	}

	if err := eg.Wait(); err != nil {
		return bencherror.FinalCheckErrs.AddError(bencherror.NewCriticalError(err, "予約情報と決済情報で不整合を検出しました"))
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
		time.Now(),
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
