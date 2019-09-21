package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chibiegg/isucon9-final/bench/assets"
	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/endpoint"
	"github.com/chibiegg/isucon9-final/bench/internal/logger"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/mock"
	"github.com/chibiegg/isucon9-final/bench/payment"
	"github.com/chibiegg/isucon9-final/bench/scenario"
	"github.com/jarcoal/httpmock"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

var (
	paymentURI, targetURI string
	assetDir              string
	messageLimit          int
)

var (
	// デバッグフラグ
	debug bool
)

// UniqueMsgs は重複除去したメッセージ配列を返します
func uniqueMsgs(msgs []string) (uniqMsgs []string) {
	dedup := map[string]struct{}{}
	for _, msg := range msgs {
		if _, ok := dedup[msg]; ok {
			continue
		}
		dedup[msg] = struct{}{}
		msgs = append(uniqMsgs, msg)
	}
	return
}

func dumpFailedResult(messages []string) {
	lgr := zap.S()

	b, err := json.Marshal(map[string]interface{}{
		"pass":     false,
		"score":    0,
		"messages": messages,
	})
	if err != nil {
		lgr.Warnf("FAILEDな結果を書き出す際にエラーが発生. messagesが失われました: %+v", err)
		fmt.Println(`{"pass": false, "score": 0, "messages": []}`)
	}

	fmt.Println(string(b))
}

var run = cli.Command{
	Name:  "run",
	Usage: "ベンチマーク実行",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:        "debug",
			Destination: &debug,
			EnvVar:      "BENCH_DEBUG",
		},
		cli.StringFlag{
			Name:        "payment",
			Value:       "http://localhost:5000",
			Destination: &paymentURI,
			EnvVar:      "BENCH_PAYMENT_URL",
		},
		cli.StringFlag{
			Name:        "target",
			Value:       "http://localhost",
			Destination: &targetURI,
			EnvVar:      "BENCH_TARGET_URL",
		},
		cli.StringFlag{
			Name:        "assetdir",
			Value:       "assets/testdata",
			Destination: &assetDir,
			EnvVar:      "BENCH_ASSETDIR",
		},
		cli.IntFlag{
			Name:        "message-limit",
			Value:       10,
			Destination: &messageLimit,
			EnvVar:      "BENCH_MESSAGE_LIMIT",
		},
	},
	Action: func(cliCtx *cli.Context) error {
		ctx := context.Background()

		lgr, err := logger.InitZapLogger()
		if err != nil {
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 1)
		}

		assets, err := assets.Load(assetDir)
		if err != nil {
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 1)
		}

		initClient, err := isutrain.NewClientForInitialize(targetURI)
		if err != nil {
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 1)
		}

		testClient, err := isutrain.NewClient(targetURI)
		if err != nil {
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 1)
		}

		paymentClient, err := payment.NewClient(paymentURI)
		if err != nil {
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 1)
		}

		if debug {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			mock.Register()
			initClient.ReplaceMockTransport()
			testClient.ReplaceMockTransport()
		}

		// initialize
		if err := paymentClient.Initialize(); err != nil {
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 0)
		}
		initClient.Initialize(ctx)
		if bencherror.InitializeErrs.IsError() {
			dumpFailedResult(bencherror.InitializeErrs.Msgs)
			return cli.NewExitError(fmt.Errorf("Initializeに失敗しました"), 0)
		}

		// pretest (まず、正しく動作できているかチェック. エラーが見つかったら、採点しようがないのでFAILにする)

		scenario.Pretest(ctx, testClient, assets)
		if bencherror.PreTestErrs.IsError() {
			dumpFailedResult(bencherror.PreTestErrs.Msgs)
			return cli.NewExitError(fmt.Errorf("Pretestに失敗しました"), 0)
		}

		// bench (ISUCOIN売り上げ計上と、減点カウントを行う)
		benchCtx, cancel := context.WithTimeout(context.Background(), config.BenchmarkTimeout)
		defer cancel()

		benchmarker := newBenchmarker(targetURI)
		benchmarker.run(benchCtx)
		if bencherror.BenchmarkErrs.IsFailure() {
			dumpFailedResult(uniqueMsgs(bencherror.BenchmarkErrs.Msgs)[:messageLimit])
			return cli.NewExitError(fmt.Errorf("Benchmarkに失敗しました"), 0)
		}

		// posttest (ベンチ後の整合性チェックにより、減点カウントを行う)
		// FIXME: 課金用のクライアントを作り、それを渡す様に変更
		score := endpoint.CalcFinalScore()

		lgr.Infof("最終チェックによるスコア: %d", score)

		// エラーカウントから、スコアを減点
		score -= bencherror.BenchmarkErrs.Penalty()

		lgr.Infof("payment   = %s", paymentURI)
		lgr.Infof("target    = %s", targetURI)
		lgr.Infof("assetdir  = %s", assetDir)

		// 最終結果をstdoutへ書き出す
		resultBytes, err := json.Marshal(map[string]interface{}{
			"pass":     true,
			"score":    score,
			"messages": uniqueMsgs(bencherror.BenchmarkErrs.Msgs)[:messageLimit],
		})
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		fmt.Println(string(resultBytes))

		return nil
	},
}
