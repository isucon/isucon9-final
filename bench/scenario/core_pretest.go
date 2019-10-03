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

	user := &isutrain.User{
		Email:    "hoge@example.com",
		Password: "hoge",
	}

	if err := client.Signup(ctx, user.Email, user.Password); err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}

	if err := client.Login(ctx, user.Email, user.Password); err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}

	_, err := client.ListStations(ctx)
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
		carNum             = 9
		seatClass          = "premium"
		adult, child       = 1, 1
	)
	_, err = client.SearchTrains(ctx, useAt, departure, arrival, "最速")
	if err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}

	// FIXME: 日付、列車クラス、名前、車両番号、乗車駅降車駅を指定
	searchTrainSeatsResp, err := client.SearchTrainSeats(ctx,
		useAt,
		trainClass, trainName, carNum, departure, arrival)
	if err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}

	availSeats := filterTrainSeats(searchTrainSeatsResp, 2)

	reserveResp, err := client.Reserve(ctx, trainClass, trainName, seatClass, availSeats, departure, arrival,
		useAt,
		carNum, child, adult)
	if err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}

	cardToken, err := paymentClient.RegistCard(ctx, "11111111", "222", "10/50")
	if err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}

	if err = client.CommitReservation(ctx, reserveResp.ReservationID, cardToken); err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}

	if err := client.CancelReservation(ctx, reserveResp.ReservationID); err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}
}

func pretestSearchTrainsNotEmptyResult(ctx context.Context, client *isutrain.Client) {
	// 初期状態で、いくつか試す
	// 必ずこれは空にならないというパターンを試す
}

func pretestSearchTrainsTimeTable(ctx context.Context, client *isutrain.Client) {
	// タイムテーブルをちゃんと返しているか試す
}

func pretestSearchTrainSeats(ctx context.Context, client *isutrain.Client) {

}

func pretestSeatAvailability(ctx context.Context, client *isutrain.Client) {

}

func pretestCheckAvailableDate(ctx context.Context, client *isutrain.Client) {
	// https://github.com/chibiegg/isucon9-final/blob/master/webapp/go/utils.go#L8
}

func pretestListStations(ctx context.Context, client *isutrain.Client) {
	// 固定でちゃんと返せてるか
}

// 異常系

// PreTestAbnormalLogin は不正なパスワードでのログインを試みます
func pretestAbnormalLogin(ctx context.Context, client *isutrain.Client) {
	if err := client.Login(ctx, "FikyavwocZear@example.com", "jieldirAwsabyonsInd", isutrain.StatusCodeOpt(http.StatusForbidden)); err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}
}
