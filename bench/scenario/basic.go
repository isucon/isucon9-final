package scenario

import (
	"context"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/score"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"golang.org/x/xerrors"
)

// BasicScenario は基本的な予約フローのシナリオです
type BasicScenario struct {
	client   *isutrain.Isutrain
	resultCh chan *RequestResult
}

func NewBasicScenario(baseURL string) *BasicScenario {
	return &BasicScenario{
		client:   isutrain.NewIsutrain(baseURL),
		resultCh: make(chan *RequestResult),
	}
}

func (s *BasicScenario) RequestResultStream() <-chan *RequestResult {
	return s.resultCh
}

func (s *BasicScenario) Run(ctx context.Context) error {
	defer close(s.resultCh)

	_, err := s.client.Register(ctx, "hoge", "hoge")
	err = xerrors.Errorf("ユーザ登録ができません: %w", err)
	SendRequestResultOrDone(ctx, s.resultCh, NewRequestResult(score.RegisterType, err))
	if err != nil {
		return err
	}

	_, err = s.client.Login(ctx, "hoge", "hoge")
	err = xerrors.Errorf("ユーザログインができません: %w", err)
	SendRequestResultOrDone(ctx, s.resultCh, NewRequestResult(score.LoginType, err))
	if err != nil {
		return err
	}

	_, err = s.client.SearchTrains(ctx, time.Now(), "東京", "大阪")
	err = xerrors.Errorf("列車検索ができません: %w", err)
	SendRequestResultOrDone(ctx, s.resultCh, NewRequestResult(score.SearchTrainsType, err))
	if err != nil {
		return err
	}

	_, err = s.client.ListTrainSeats(ctx, "こだま", "96号")
	err = xerrors.Errorf("列車の座席座席列挙できません: %w", err)
	SendRequestResultOrDone(ctx, s.resultCh, NewRequestResult(score.ListTrainSeatsType, err))
	if err != nil {
		return err
	}

	_, err = s.client.Reserve(ctx)
	err = xerrors.Errorf("予約ができません: %w", err)
	SendRequestResultOrDone(ctx, s.resultCh, NewRequestResult(score.ReserveType, err))
	if err != nil {
		return err
	}

	_, err = s.client.CommitReservation(ctx, 1111)
	err = xerrors.Errorf("予約を確定できませんでした: %w", err)
	SendRequestResultOrDone(ctx, s.resultCh, NewRequestResult(score.CommitReservationType, err))
	if err != nil {
		return err
	}

	_, err = s.client.ListReservations(ctx)
	err = xerrors.Errorf("予約を列挙できませんでした: %w", err)
	SendRequestResultOrDone(ctx, s.resultCh, NewRequestResult(score.ListReservationsType, err))
	if err != nil {
		return err
	}

	return nil
}
