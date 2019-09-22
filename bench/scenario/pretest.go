package scenario

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/chibiegg/isucon9-final/bench/assets"
	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"golang.org/x/sync/errgroup"
)

// Guest訪問
var (
	ErrInitialTrainDatasetCount = errors.New("列車初期データセットの件数が一致しません")
	ErrInvalidAssetHash         = errors.New("静的ファイルの内容が正しくありません")
)

// Pretest は、ベンチマーク前のアプリケーションが正常に動作できているか検証し、できていなければFAILとします
func Pretest(ctx context.Context, client *isutrain.Client, assets []*assets.Asset) {
	pretestGrp, _ := errgroup.WithContext(ctx)

	// 静的ファイル
	pretestGrp.Go(func() error {
		pretestStaticFiles(ctx, client, assets)
		return nil
	})
	// 正常系
	pretestGrp.Go(func() error {
		pretestNormalReservation(ctx, client)
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

		if hash != asset.Hash {
			return bencherror.PreTestErrs.AddError(bencherror.NewApplicationError(ErrInvalidAssetHash, "GET /%s: 静的ファイルのハッシュ値が異なります", asset.Path))
		}
	}
	return nil
}

// 正常系

// preTestNormalReservation は予約までの一連の流れを検証します
func pretestNormalReservation(ctx context.Context, client *isutrain.Client) {
	// FIXME: 最初から登録されて入る、２つくらいのユーザで試す
	if err := client.Register(ctx, "hoge", "hoge", nil); err != nil {
		bencherror.PreTestErrs.AddError(bencherror.NewCriticalError(err, "ユーザ登録に失敗しました: user=%s, password=%s", "hoge", "hoge"))
		return
	}

	if err := client.Login(ctx, "hoge", "hoge", nil); err != nil {
		bencherror.PreTestErrs.AddError(bencherror.NewCriticalError(err, "ユーザのログインに失敗しました: user=%s, password=%s", "hoge", "hoge"))
		return
	}

	_, err := client.ListStations(ctx, nil)
	if err != nil {
		bencherror.PreTestErrs.AddError(bencherror.NewCriticalError(err, "駅一覧を取得できません"))
		return
	}

	_, err = client.SearchTrains(ctx, time.Now(), "東京", "大阪", nil)
	if err != nil {
		bencherror.PreTestErrs.AddError(bencherror.NewCriticalError(err, "列車検索ができません"))
		return
	}

	_, err = client.ListTrainSeats(ctx, "こだま", "96号", nil)
	if err != nil {
		bencherror.PreTestErrs.AddError(bencherror.NewCriticalError(err, "列車の座席座席列挙できません"))
		return
	}

	reservation, err := client.Reserve(ctx, nil)
	if err != nil {
		bencherror.PreTestErrs.AddError(bencherror.NewCriticalError(err, "予約ができません"))
		return
	}

	if err = client.CommitReservation(ctx, reservation.ReservationID, nil); err != nil {
		bencherror.PreTestErrs.AddError(bencherror.NewCriticalError(err, "予約を確定できませんでした"))
		return
	}

	if err := client.CancelReservation(ctx, reservation.ReservationID, nil); err != nil {
		bencherror.PreTestErrs.AddError(bencherror.NewCriticalError(err, "予約をキャンセルできませんでした"))
		return
	}
}

// PreTestNormalSearch は検索条件を細かく指定して検索します
func pretestNormalSearch(ctx context.Context, client *isutrain.Client) {
}

// 異常系

// PreTestAbnormalLogin は不正なパスワードでのログインを試みます
func pretestAbnormalLogin(ctx context.Context, client *isutrain.Client) {
	if err := client.Login(ctx, "username", "password", nil); err != nil {

	}
}
