package scenario

import (
	"context"
	"fmt"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/xrandom"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

// BasicScenario は基本的な予約フローのシナリオです
type BasicScenario struct {
	Client *isutrain.Client
}

func NewBasicScenario(targetBaseURL string) (*BasicScenario, error) {
	client, err := isutrain.NewClient(targetBaseURL)
	if err != nil {
		return nil, err
	}

	return &BasicScenario{
		Client: client,
	}, nil
}

// TODO: シナリオ１回転で同期ポイント
func (s *BasicScenario) Run(ctx context.Context) error {
	err := s.Client.Register(ctx, "hoge", "hoge", nil)
	if err != nil {
		benchErr := fmt.Errorf("ユーザ登録ができません: %w", err)
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	err = s.Client.Login(ctx, "hoge", "hoge", nil)
	if err != nil {
		benchErr := fmt.Errorf("ユーザログインができません: %w", err)
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	_, err = s.Client.ListStations(ctx, nil)
	if err != nil {
		benchErr := fmt.Errorf("駅一覧を取得できません: %w", err)
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	var (
		origin      = xrandom.GetRandomStations()
		destination = xrandom.GetRandomStations()
	)
	_, err = s.Client.SearchTrains(ctx, time.Now(), origin, destination, nil)
	if err != nil {
		benchErr := fmt.Errorf("列車検索ができません: %w", err)
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	_, err = s.Client.ListTrainSeats(ctx, "こだま", "96号", nil)
	if err != nil {
		benchErr := fmt.Errorf("列車の座席座席列挙できません: %w", err)
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	reservation, err := s.Client.Reserve(ctx, nil)
	if err != nil {
		benchErr := fmt.Errorf("予約ができません: %w", err)
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	err = s.Client.CommitReservation(ctx, reservation.ReservationID, nil)
	if err != nil {
		benchErr := fmt.Errorf("予約を確定できませんでした: %w", err)
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	_, err = s.Client.ListReservations(ctx, nil)
	if err != nil {
		benchErr := fmt.Errorf("予約を列挙できませんでした: %w", err)
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	return nil
}
