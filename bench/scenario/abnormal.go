package scenario

import "context"

func AbnormalLoginScenario(ctx context.Context) error {
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
