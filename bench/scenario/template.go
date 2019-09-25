package scenario

import (
	"context"
	"go.uber.org/zap"
	"log"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/xrandom"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/payment"
)

func AwesomeScenario(ctx context.Context) error {
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
	if config.Debug {
		client.ReplaceMockTransport()
	}

	// ここからシナリオ

	// 運営だけが見る情報は log に投げる
	log.Printf("[template:AwesomeScenario] これは素晴らしいシナリオです\n")

	// ユーザー作成とログイン
	user, err := xrandom.GetRandomUser() // ランダムデータ生成系は xrandom に作成するかあるものを使う
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.Signup(ctx, user.Email, user.Password, nil)
	if err != nil {
		// シナリオに於けるユーザーに見せるメッセージは `lgr.Infow()` に送る
		lgr.Infow("ユーザ登録失敗",
			"error", err,
			"email", user.Email,
			"password", user.Password,
		)
		// `bencherror.BenchmarkErrs.AddError(err)` も忘れずに
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.Login(ctx, user.Email, user.Password, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	cardToken, err := paymentClient.RegistCard(ctx, "11111111", "222", "10/50")
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	log.Printf("[template:AwesomeScenario] カードのトークン %s\n", cardToken)

	return nil

}
