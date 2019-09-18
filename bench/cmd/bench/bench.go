package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/logger"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/mock"
	"github.com/chibiegg/isucon9-final/bench/scenario"
	"github.com/jarcoal/httpmock"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

var (
	paymentURI, targetURI string
	dataDir, staticDir    string
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
		},
		cli.StringFlag{
			Name:        "payment",
			Value:       "http://localhost:5000",
			Destination: &paymentURI,
		},
		cli.StringFlag{
			Name:        "target",
			Value:       "http://localhost",
			Destination: &targetURI,
		},
		cli.StringFlag{
			Name:        "datadir",
			Value:       "/home/isucon/isucon9-final/data",
			Destination: &dataDir,
		},
		cli.StringFlag{
			Name:        "staticdir",
			Value:       "/home/isucon/isucon9-final/static",
			Destination: &staticDir,
		},
	},
	Action: func(cliCtx *cli.Context) error {
		lgr, err := logger.InitZapLogger()
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

		if debug {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			mock.Register()
			initClient.ReplaceMockTransport()
			testClient.ReplaceMockTransport()
		}

		// TODO: 初期データのロードなど用意

		// initialize
		initClient.Initialize(context.Background())
		if bencherror.InitializeErrs.IsError() {
			dumpFailedResult(bencherror.InitializeErrs.Msgs)
			return cli.NewExitError(fmt.Errorf("Initializeに失敗しました"), 1)
		}

		// pretest (まず、正しく動作できているかチェック. エラーが見つかったら、採点しようがないのでFAILにする)

		scenario.Pretest(testClient, staticDir)
		if bencherror.PreTestErrs.IsError() {
			dumpFailedResult(bencherror.PreTestErrs.Msgs)
			return cli.NewExitError(fmt.Errorf("Pretestに失敗しました"), 1)
		}

		// bench (ISUCOIN売り上げ計上と、減点カウントを行う)
		ctx, cancel := context.WithTimeout(context.Background(), config.BenchmarkTimeout)
		defer cancel()

		benchmarker := newBenchmarker(targetURI, debug)
		benchmarker.run(ctx)
		if bencherror.BenchmarkErrs.IsFailure() {
			dumpFailedResult(bencherror.BenchmarkErrs.Msgs)
			return cli.NewExitError(fmt.Errorf("Benchmarkに失敗しました"), 1)
		}

		// posttest (ベンチ後の整合性チェックにより、減点カウントを行う)
		// FIXME: 課金用のクライアントを作り、それを渡す様に変更
		score := scenario.FinalCheck(testClient)

		// エラーカウントから、スコアを減点
		score -= bencherror.BenchmarkErrs.Penalty()

		lgr.Infof("payment   = %s", paymentURI)
		lgr.Infof("target    = %s", targetURI)
		lgr.Infof("datadir   = %s", dataDir)
		lgr.Infof("staticdir = %s", staticDir)

		// 最終結果をstdoutへ書き出す
		resultBytes, err := json.Marshal(map[string]interface{}{
			"pass":     true,
			"score":    score,
			"messages": bencherror.BenchmarkErrs.Msgs,
		})
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		fmt.Println(string(resultBytes))

		return nil
	},
}
