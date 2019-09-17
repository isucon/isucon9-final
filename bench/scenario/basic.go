package scenario

import (
	"context"
	"fmt"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

// BasicScenario は基本的な予約フローのシナリオです
type BasicScenario struct {
	client *isutrain.Client
}

func NewBasicScenario(targetBaseURL string) (*BasicScenario, error) {
	client, err := isutrain.NewClient(targetBaseURL)
	if err != nil {
		return nil, err
	}

	return &BasicScenario{
		client: client,
	}, nil
}

// TODO: シナリオ１回転で同期ポイント
func (s *BasicScenario) Run(ctx context.Context) error {
	err := s.client.Register(ctx, "hoge", "hoge")
	if err != nil {
		benchErr := fmt.Errorf("ユーザ登録ができません: %w", err)
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	err = s.client.Login(ctx, "hoge", "hoge")
	if err != nil {
		benchErr := fmt.Errorf("ユーザログインができません: %w", err)
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	_, err = s.client.ListStations(ctx)
	if err != nil {
		benchErr := fmt.Errorf("駅一覧を取得できません: %w", err)
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	_, err = s.client.SearchTrains(ctx, time.Now(), "東京", "大阪")
	if err != nil {
		benchErr := fmt.Errorf("列車検索ができません: %w", err)
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	_, err = s.client.ListTrainSeats(ctx, "こだま", "96号")
	if err != nil {
		benchErr := fmt.Errorf("列車の座席座席列挙できません: %w", err)
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	reservation, err := s.client.Reserve(ctx)
	if err != nil {
		benchErr := fmt.Errorf("予約ができません: %w", err)
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	err = s.client.CommitReservation(ctx, reservation.ReservationID)
	if err != nil {
		benchErr := fmt.Errorf("予約を確定できませんでした: %w", err)
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	_, err = s.client.ListReservations(ctx)
	if err != nil {
		benchErr := fmt.Errorf("予約を列挙できませんでした: %w", err)
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	return nil
}
