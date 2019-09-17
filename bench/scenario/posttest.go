package scenario

import (
	"errors"

	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

var (
	ErrListReservation = errors.New("予約一覧の取得に失敗しました")
)

func PostTest(client *isutrain.Client) error {
	return nil
}
