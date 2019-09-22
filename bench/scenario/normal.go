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
func NormalScenario(ctx context.Context) error {
	client, err := isutrain.NewClient()
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
		departure = xrandom.GetRandomStations()
		arrival   = xrandom.GetRandomStations()
	)
	_, err = client.SearchTrains(ctx, time.Now(), departure, arrival, nil)
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
func NormalCancelScenario(ctx context.Context) error {
	client, err := isutrain.NewClient()
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
		departure = xrandom.GetRandomStations()
		arrival   = xrandom.GetRandomStations()
	)
	_, err = client.SearchTrains(ctx, time.Now(), departure, arrival, nil)
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

	err = client.CancelReservation(ctx, reservation.ReservationID, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewApplicationError(err, "予約をキャンセルできませんでした"))
	}

	return nil
}

// 曖昧検索シナリオ
func NormalAmbigiousSearchScenario(ctx context.Context) error {
	return nil
}

// オリンピック開催期間に負荷をあげるシナリオ
// FIXME: initializeで指定された日数に応じ、負荷レベルを変化させる
func NormalOlympicParticipantsScenario(ctx context.Context) error {

	// NOTE: webappから見て、明らかに負荷が上がったと感じるレベルに持ってく必要がある
	// NOTE: 指定できる最大の日数で負荷をかける際、飽和しないようにする
	// NOTE: 

	return nil
}
