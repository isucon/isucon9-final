package scenario

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/xrandom"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/payment"
)

func DoSomething() (int, error) {
	return -1, fmt.Errorf("always error")
}

func AwesomeScenario(ctx context.Context) error {
	// zapロガー取得 (参考: https://qiita.com/emonuh/items/28dbee9bf2fe51d28153#sugaredlogger )
	lgr := zap.S()

	// ISUTRAIN APIのクライアントを作成
	client, err := isutrain.NewClient()
	if err != nil {
		// 実行中のエラーは `bencherror.BenchmarkErrs.AddError(err)` に投げる
		return bencherror.BenchmarkErrs.AddError(err)
	}

	// 決済サービスのクライアントを作成
	paymentClient, err := payment.NewClient()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	// デバッグの場合はモックに差し替える
	// NOTE: httpmockというライブラリが、http.Transporterを差し替えてエンドポイントをマウントする都合上、この処理が必要です
	//       この処理がないと、テスト実行時に存在しない宛先にリクエストを送り、失敗します
	if config.Debug {
		client.ReplaceMockTransport()
	}

	// ここからシナリオ

	// 運営だけが見る情報は zap に投げる
	lgr.Info("[template:AwesomeScenario] これは素晴らしいシナリオです")

	// ユーザー作成とログイン
	user, err := xrandom.GetRandomUser() // ランダムデータ生成系は xrandom に作成するかあるものを使う
	if err != nil {
		bencherror.SystemErrs.AddError(err)
		return nil
	}

	err = client.Signup(ctx, user.Email, user.Password)
	if err != nil {
		// `bencherror.BenchmarkErrs.AddError(err)` も忘れずに
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.Login(ctx, user.Email, user.Password)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	cardToken, err := paymentClient.RegistCard(ctx, "11111111", "222", "10/50")
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	// 例) cardTokenが不正な場合に失格にしたい場合
	if cardToken != "XXXXXXXX" {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewSimpleCriticalError("カードトークンが不正: %s", cardToken))
	}
	// 例) Client以外からエラーを得たが、これを減点要素にしたい場合
	if num, err := DoSomething(); err != nil {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewApplicationError(err, "エラー発生: %d", num))
	}

	lgr.Infof("[template:AwesomeScenario] カードのトークン %s", cardToken)

	return nil

}
