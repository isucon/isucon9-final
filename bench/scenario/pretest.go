package scenario

import (
	"errors"

	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
)

// Guest訪問
var (
	ErrInitialTrainDatasetCount = errors.New("列車初期データセットの件数が一致しません")
)

// PreTest は、ベンチマーク前のアプリケーションが正常に動作できているか検証し、できていなければFAILとします
func PreTest(client *isutrain.Client) {
	// 正常系
	if err := preTestNormalReservation(client); err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}
	if err := preTestNormalCancelReservation(client); err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}
	if err := preTestNormalSearch(client); err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}
	// 異常系
	if err := preTestAbnormalLogin(client); err != nil {
		bencherror.PreTestErrs.AddError(err)
		return
	}
}

// 正常系

// preTestNormalReservation は予約までの一連の流れを検証します
func preTestNormalReservation(client *isutrain.Client) error {

	return nil
}

// PreTestNormalCancelReservation は予約を行ったのち、キャンセルして一覧チェックするまでの一連の流れを検証します
func preTestNormalCancelReservation(client *isutrain.Client) error {
	return nil
}

// PreTestNormalSearch は検索条件を細かく指定して検索します
func preTestNormalSearch(client *isutrain.Client) error {
	return nil
}

// 異常系

// PreTestAbnormalLogin は不正なパスワードでのログインを試みます
func preTestAbnormalLogin(client *isutrain.Client) error {
	return nil
}

