package scenario

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"time"

	"github.com/chibiegg/isucon9-final/bench/assets"
	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/payment"
	"golang.org/x/sync/errgroup"
)

// Guest訪問
var (
	ErrInitialTrainDatasetCount = errors.New("列車初期データセットの件数が一致しません")
	ErrInvalidAssetHash         = errors.New("静的ファイルの内容が正しくありません")
)

// Pretest は、ベンチマーク前のアプリケーションが正常に動作できているか検証し、できていなければFAILとします
func Pretest(ctx context.Context, client *isutrain.Client, paymentClient *payment.Client, assets []*assets.Asset) {
	pretestGrp, _ := errgroup.WithContext(ctx)

	// 静的ファイル
	pretestGrp.Go(func() error {
		pretestStaticFiles(ctx, client, assets)
		return nil
	})
	// 正常系
	pretestGrp.Go(func() error {
		pretestNormalReservation(ctx, client, paymentClient)
		return nil
	})
	pretestGrp.Go(func() error {
		pretestNormalSearch(ctx, client)
		return nil
	})
	// 異常系
	pretestGrp.Go(func() error {
		pretestAbnormalLogin(ctx, client)
		return nil
	})

	pretestGrp.Wait()
}

// 静的ファイルチェック
func pretestStaticFiles(ctx context.Context, client *isutrain.Client, assets []*assets.Asset) error {
	for _, asset := range assets {
		b, err := client.DownloadAsset(ctx, asset.Path)
		if err != nil {
			return err
		}

		hash := sha256.Sum256(b)

		if !bytes.Equal(hash[:], asset.Hash[:]) {
			return bencherror.PreTestErrs.AddError(bencherror.NewApplicationError(ErrInvalidAssetHash, "GET /%s: 静的ファイルのハッシュ値が異なります", asset.Path))
		}
	}
	return nil
}

// 正常系

// preTestNormalReservation は予約までの一連の流れを検証します
func pretestNormalReservation(ctx context.Context, client *isutrain.Client, paymentClient *payment.Client) {
	// FIXME: 最初から登録されて入る、２つくらいのユーザで試す

	if err := client.Signup(ctx, "hoge@example.com", "hoge", nil); err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}

	if err := client.Login(ctx, "hoge@example.com", "hoge", nil); err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}

	_, err := client.ListStations(ctx, nil)
	if err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}

	// FIXME: 区間を考慮しつつ、駅名をランダムに取得する必要がある
	// ここで
	// * 列車クラス、名前
	// * 列車の発駅着駅
	// * 利用者の乗車駅降車駅
	// * 利用者の乗車時刻 降車時刻
	// * 空席情報 (プレミアムにできる？プレミアムかつ喫煙にできる？そもそも予約できる？予約できてかつ喫煙できる？未予約？)
	// * 空席情報それぞれの料金
	// を得られる
	var (
		useAt              = time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)
		departure, arrival = "東京", "大阪"
		trainClass         = "最速"
		trainName          = "49"
		carNum             = 8
		seatClass          = "premium"
		adult, child       = 1, 1
		seatType           = ""
	)
	_, err = client.SearchTrains(ctx, useAt, departure, arrival, "", nil)
	if err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}

	// FIXME: 日付、列車クラス、名前、車両番号、乗車駅降車駅を指定
	seatsResp, err := client.ListTrainSeats(ctx,
		useAt,
		trainClass, trainName, carNum, departure, arrival, nil)
	if err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}

	validSeats, err := assertListTrainSeats(seatsResp, 2)
	if err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}

	_, reservation, err := client.Reserve(ctx, trainClass, trainName, seatClass, validSeats, departure, arrival,
		useAt,
		carNum, adult, child, seatType, nil)
	if err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}

	// TODO: assertReserve

	cardToken, err := paymentClient.RegistCard(ctx, "11111111", "222", "10/50")
	if err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}

	if err = client.CommitReservation(ctx, reservation.ReservationID, cardToken, nil); err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}

	if err := client.CancelReservation(ctx, reservation.ReservationID, nil); err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}
}

// PreTestNormalSearch は検索条件を細かく指定して検索します
func pretestNormalSearch(ctx context.Context, client *isutrain.Client) {
}

// 異常系

// PreTestAbnormalLogin は不正なパスワードでのログインを試みます
func pretestAbnormalLogin(ctx context.Context, client *isutrain.Client) {
	if err := client.Login(ctx, "username", "password", &isutrain.ClientOption{
		WantStatusCode: http.StatusForbidden,
	}); err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}
}
