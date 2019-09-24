package scenario

import (
	"context"
	"net/http"

	"github.com/chibiegg/isucon9-final/bench/internal/util"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

func AbnormalLoginScenario(ctx context.Context) error {
	var (
		email, err1    = util.SecureRandomStr(10)
		password, err2 = util.SecureRandomStr(10)
	)
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	client, err := isutrain.NewClient()
	if err != nil {
		return err
	}

	err = client.Login(ctx, email, password, &isutrain.ClientOption{
		WantStatusCode: http.StatusUnauthorized,
	})
	if err != nil {
		return err
	}

	return nil
}

// 指定列車の運用区間外で予約を取ろうとして、きちんと弾かれるかチェック
func AbnormalReserveWrongSection() {
}

// 列車の指定号車に存在しない席を予約しようとし、エラーになるかチェック
func AbnormalReserveWrongSeat() {

}

func AbnormalReserveWithCSRFTokenScenario(ctx context.Context) error {
	return nil
}
