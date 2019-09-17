package scenario

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"golang.org/x/sync/errgroup"
)

// Guest訪問
var (
	ErrInitialTrainDatasetCount = errors.New("列車初期データセットの件数が一致しません")
)

// PreTest は、ベンチマーク前のアプリケーションが正常に動作できているか検証し、できていなければFAILとします
func Pretest(client *isutrain.Client) {
	pretestGrp, _ := errgroup.WithContext(context.Background())

	// 静的ファイル
	pretestGrp.Go(func() error {
		pretestStaticFiles()
		return nil
	})
	// 正常系
	pretestGrp.Go(func() error {
		pretestNormalReservation(client)
		return nil
	})
	pretestGrp.Go(func() error {
		pretestNormalSearch(client)
		return nil
	})
	// 異常系
	pretestGrp.Go(func() error {
		pretestAbnormalLogin(client)
		return nil
	})

	pretestGrp.Wait()
	return
}

// 静的ファイル
func pretestStaticFiles() error {
	return nil
}

// 正常系

// preTestNormalReservation は予約までの一連の流れを検証します
func pretestNormalReservation(client *isutrain.Client) {
	ctx := context.Background()

	if err := client.Register(ctx, "hoge", "hoge"); err != nil {
		pretestErr := fmt.Errorf("ユーザ登録ができません: %w", err)
		bencherror.PreTestErrs.AddError(pretestErr)
		return
	}

	if err := client.Login(ctx, "hoge", "hoge"); err != nil {
		pretestErr := fmt.Errorf("ユーザログインができません: %w", err)
		bencherror.PreTestErrs.AddError(pretestErr)
		return
	}

	_, err := client.ListStations(ctx)
	if err != nil {
		pretestErr := fmt.Errorf("駅一覧を取得できません: %w", err)
		bencherror.PreTestErrs.AddError(pretestErr)
		return
	}

	_, err = client.SearchTrains(ctx, time.Now(), "東京", "大阪")
	if err != nil {
		pretestErr := fmt.Errorf("列車検索ができません: %w", err)
		bencherror.PreTestErrs.AddError(pretestErr)
		return
	}

	_, err = client.ListTrainSeats(ctx, "こだま", "96号")
	if err != nil {
		pretestErr := fmt.Errorf("列車の座席座席列挙できません: %w", err)
		bencherror.PreTestErrs.AddError(pretestErr)
		return
	}

	reservation, err := client.Reserve(ctx)
	if err != nil {
		pretestErr := fmt.Errorf("予約ができません: %w", err)
		bencherror.PreTestErrs.AddError(pretestErr)
		return
	}

	if err = client.CommitReservation(ctx, reservation.ReservationID); err != nil {
		pretestErr := fmt.Errorf("予約を確定できませんでした: %w", err)
		bencherror.PreTestErrs.AddError(pretestErr)
		return
	}

	if err := client.CancelReservation(ctx, reservation.ReservationID); err != nil {
		pretestErr := fmt.Errorf("予約をキャンセルできませんでした: %w", err)
		bencherror.PreTestErrs.AddError(pretestErr)
		return
	}
}

// PreTestNormalSearch は検索条件を細かく指定して検索します
func pretestNormalSearch(client *isutrain.Client) {
}

// 異常系

// PreTestAbnormalLogin は不正なパスワードでのログインを試みます
func pretestAbnormalLogin(client *isutrain.Client) {
	ctx := context.Background()

	if err := client.Login(ctx, "username", "password"); err != nil {

	}
}
