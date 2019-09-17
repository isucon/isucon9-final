package scenario

import (
	"errors"

	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

var (
	ErrListReservation = errors.New("予約一覧の取得に失敗しました")
)

// FinalCheck は、課金サービスとwebappとで決済情報を突き合わせ、売上を計上します
func FinalCheck(client *isutrain.Client) uint64 {
	finalcheckPayment()
	return 0
}

// PosttestPayment は課金情報の整合性チェックを行います
func finalcheckPayment() {
	// FIXME: 課金とのつなぎこみ
}
