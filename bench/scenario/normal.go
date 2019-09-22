package scenario

import (
	"context"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/xrandom"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

// NormalScenario は基本的な予約フローのシナリオです
func NormalScenario(ctx context.Context, targetBaseURL string) error {
	client, err := isutrain.NewClient(targetBaseURL)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewCriticalError(err, "Isutrainクライアントが作成できません. 運営に確認をお願いいたします"))
	}

	if config.Debug {
		client.ReplaceMockTransport()
	}

	err = client.Register(ctx, "hoge", "hoge", nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewApplicationError(err, "ユーザ登録ができません"))
	}

	err = client.Login(ctx, "hoge", "hoge", nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewApplicationError(err, "ユーザログインができません"))
	}

	_, err = client.ListStations(ctx, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewApplicationError(err, "駅一覧を取得できません"))
	}

	var (
		origin      = xrandom.GetRandomStations()
		destination = xrandom.GetRandomStations()
	)
	_, err = client.SearchTrains(ctx, time.Now(), origin, destination, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewApplicationError(err, "列車検索ができません"))
	}

	_, err = client.ListTrainSeats(ctx, "こだま", "96号", 1, "東京", "大阪", nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewApplicationError(err, "列車の座席座席列挙できません"))
	}

	reservation, err := client.Reserve(ctx, "こだま", "69号", "premium", isutrain.TrainSeats{}, "東京", "名古屋", time.Now(), 1, 1, 1, "isle", nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewApplicationError(err, "予約ができません"))
	}

	err = client.CommitReservation(ctx, reservation.ReservationID, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewApplicationError(err, "予約を確定できませんでした"))
	}

	_, err = client.ListReservations(ctx, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewApplicationError(err, "予約を列挙できませんでした"))
	}

	return nil
}

// 予約キャンセル含む(Commit後にキャンセル)
func NormalCancelScenario(ctx context.Context, targetBaseURL string) error {

	return nil
}

// 曖昧検索シナリオ
func NormalAmbigiousSearchScenario(ctx context.Context, targetBaseURL string) error {

	return nil
}
