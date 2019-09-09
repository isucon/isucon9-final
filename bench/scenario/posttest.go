package scenario

import "errors"

var (
	ErrListReservation = errors.New("予約一覧の取得に失敗しました")
)

func PostTest() error {
	return nil
}
