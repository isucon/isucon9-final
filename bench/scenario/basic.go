package scenario

import (
	"context"
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
		benchErr := bencherror.NewApplicationError(err, "ユーザ登録ができません")
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	err = s.Client.Login(ctx, "hoge", "hoge", nil)
	if err != nil {
		benchErr := bencherror.NewApplicationError(err, "ユーザログインができません")
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	_, err = s.Client.ListStations(ctx, nil)
	if err != nil {
		benchErr := bencherror.NewApplicationError(err, "駅一覧を取得できません")
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	var (
		origin      = xrandom.GetRandomStations()
		destination = xrandom.GetRandomStations()
	)
	_, err = s.Client.SearchTrains(ctx, time.Now(), origin, destination, nil)
	if err != nil {
		benchErr := bencherror.NewApplicationError(err, "列車検索ができません")
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	_, err = s.Client.ListTrainSeats(ctx, "こだま", "96号", nil)
	if err != nil {
		benchErr := bencherror.NewApplicationError(err, "列車の座席座席列挙できません")
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	reservation, err := s.Client.Reserve(ctx, nil)
	if err != nil {
		benchErr := bencherror.NewApplicationError(err, "予約ができません")
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	err = s.Client.CommitReservation(ctx, reservation.ReservationID, nil)
	if err != nil {
		benchErr := bencherror.NewApplicationError(err, "予約を確定できませんでした")
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	_, err = s.Client.ListReservations(ctx, nil)
	if err != nil {
		benchErr := bencherror.NewApplicationError(err, "予約を列挙できませんでした")
		bencherror.BenchmarkErrs.AddError(benchErr)
		return benchErr
	}

	return nil
}
