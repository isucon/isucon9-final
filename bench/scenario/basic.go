package scenario

import (
	"context"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/session"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"golang.org/x/xerrors"
)

// BasicScenario は基本的な予約フローのシナリオです
type BasicScenario struct {
	client *isutrain.Isutrain
}

func NewBasicScenario(baseURL string) *BasicScenario {
	return &BasicScenario{
		client: isutrain.NewIsutrain(baseURL),
	}
}

// FIXME: 結果やエラーの共有は、共有オブジェクトを介するように

func (s *BasicScenario) Run(ctx context.Context) error {
	sess, err := session.NewSession()
	if err != nil {
		return err
	}

	_, err = s.client.Register(ctx, sess, "hoge", "hoge")
	err = xerrors.Errorf("ユーザ登録ができません: %w", err)
	if err != nil {
		return err
	}

	_, err = s.client.Login(ctx, sess, "hoge", "hoge")
	err = xerrors.Errorf("ユーザログインができません: %w", err)
	if err != nil {
		return err
	}

	_, err = s.client.SearchTrains(ctx, sess, time.Now(), "東京", "大阪")
	err = xerrors.Errorf("列車検索ができません: %w", err)
	if err != nil {
		return err
	}

	_, err = s.client.ListTrainSeats(ctx, sess, "こだま", "96号")
	err = xerrors.Errorf("列車の座席座席列挙できません: %w", err)
	if err != nil {
		return err
	}

	_, err = s.client.Reserve(ctx, sess)
	err = xerrors.Errorf("予約ができません: %w", err)
	if err != nil {
		return err
	}

	_, err = s.client.CommitReservation(ctx, sess, 1111)
	err = xerrors.Errorf("予約を確定できませんでした: %w", err)
	if err != nil {
		return err
	}

	_, err = s.client.ListReservations(ctx, sess)
	err = xerrors.Errorf("予約を列挙できませんでした: %w", err)
	if err != nil {
		return err
	}

	return nil
}
