package isutrain

import (
	"context"
	"log"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/result"
	"github.com/chibiegg/isucon9-final/bench/internal/score"
	"github.com/chibiegg/isucon9-final/bench/util"
	"golang.org/x/xerrors"
)

// Scenario isutrain のベンチマークシナリオ
type Scenario interface {
	BenchResultStream() <-chan *result.BenchResult
	Run(ctx context.Context) error
}

func NewScenario(tick uint64, baseURL string) Scenario {
	return NewBasicScenario(baseURL)
}

// BasicScenario は基本的な予約フローのシナリオです
type BasicScenario struct {
	client   *Isutrain
	resultCh chan *result.BenchResult
}

func NewBasicScenario(baseURL string) *BasicScenario {
	return &BasicScenario{
		client:   NewIsutrain(baseURL),
		resultCh: make(chan *result.BenchResult),
	}
}

func (s *BasicScenario) BenchResultStream() <-chan *result.BenchResult {
	return s.resultCh
}

func (s *BasicScenario) Run(ctx context.Context) error {
	defer close(s.resultCh)

	log.Println("1")
	_, err := s.client.Register("hoge", "hoge")
	util.DoneOrSendBenchResult(ctx, s.resultCh, result.NewBenchResult(score.RegisterType, err))
	if err != nil {
		return xerrors.Errorf("ユーザ登録ができません: %w", err)
	}

	log.Println("2")
	_, err = s.client.Login("hoge", "hoge")
	util.DoneOrSendBenchResult(ctx, s.resultCh, result.NewBenchResult(score.LoginType, err))
	if err != nil {
		return xerrors.Errorf("ユーザログインができません: %w", err)
	}

	log.Println("3")
	_, err = s.client.SearchTrains(time.Now(), "東京", "大阪")
	util.DoneOrSendBenchResult(ctx, s.resultCh, result.NewBenchResult(score.SearchTrainsType, err))
	if err != nil {
		return xerrors.Errorf("列車検索ができません: %w", err)
	}

	log.Println("4")
	_, err = s.client.ListTrainSeats("こだま", "96号")
	util.DoneOrSendBenchResult(ctx, s.resultCh, result.NewBenchResult(score.ListTrainSeatsType, err))
	if err != nil {
		return xerrors.Errorf("列車の座席座席列挙できません: %w", err)
	}

	log.Println("5")
	_, err = s.client.Reserve()
	util.DoneOrSendBenchResult(ctx, s.resultCh, result.NewBenchResult(score.ReserveType, err))
	if err != nil {
		return xerrors.Errorf("予約ができません: %w", err)
	}

	log.Println("6")
	_, err = s.client.CommitReservation(1111)
	util.DoneOrSendBenchResult(ctx, s.resultCh, result.NewBenchResult(score.CommitReservationType, err))
	if err != nil {
		return xerrors.Errorf("予約を確定できませんでした: %w", err)
	}

	log.Println("7")
	_, err = s.client.ListReservations()
	util.DoneOrSendBenchResult(ctx, s.resultCh, result.NewBenchResult(score.ListReservationsType, err))
	if err != nil {
		return xerrors.Errorf("予約を列挙できませんでした: %w", err)
	}

	return nil
}
