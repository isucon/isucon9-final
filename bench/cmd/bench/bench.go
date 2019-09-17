package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/logger"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/mock"
	"github.com/chibiegg/isucon9-final/bench/scenario"
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

func dumpFailedResult(messages []string) {
	lgr := zap.S()

	b, err := json.Marshal(map[string]interface{}{
		"pass":     false,
		"score":    0,
		"messages": messages,
	})
	if err != nil {
		lgr.Warn("FAILEDな結果を書き出す際にエラーが発生. messagesが失われました", zap.Error(err))
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
			Value:       "http://127.0.0.1",
			Destination: &paymentURI,
		},
		cli.StringFlag{
			Name:        "target",
			Value:       "http://127.0.0.1",
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

		if debug {
			m := mock.Register()
			log.Println(m.Delay())
		}

		// TODO: 初期データのロードなど用意

		// initialize
		initClient, err := isutrain.NewClientForInitialize("http://127.0.0.1:8000/")
		if err != nil {
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 1)
		}
		initClient.Initialize(context.Background())
		if err != nil {
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 1)
		}

		// pretest (まず、正しく動作できているかチェック. エラーが見つかったら、採点しようがないのでFAILにする)
		testClient, err := isutrain.NewClient("http://127.0.0.1:8000/")
		if err != nil {
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 1)
		}
		scenario.PreTest(testClient)
		if bencherror.PreTestErrs.IsError() {
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 1)
		}

		// bench (ISUCOIN売り上げ計上と、減点カウントを行う)
		// ctx, cancel := context.WithTimeout(context.Background(), config.BenchmarkTimeout)
		// defer cancel()

		// benchmarker, err := newBenchmarker("http://127.0.0.1:8000/")
		// if err != nil {
		// 	return cli.NewExitError(err, 1)
		// }
		// err = benchmarker.run(ctx)
		// if err != nil {
		// 	return cli.NewExitError(err, 1)
		// }

		// posttest (ベンチ後の整合性チェックにより、減点カウントを行う)
		scenario.PostTest(testClient)
		if bencherror.PostTestErrs.IsError() {
			return cli.NewExitError(err, 1)
		}

		lgr.Infof("payment   = %s", paymentURI)
		lgr.Infof("target    = %s", targetURI)
		lgr.Infof("datadir   = %s", dataDir)
		lgr.Infof("staticdir = %s", staticDir)

		// 最終結果をstdoutへ書き出す
		resultBytes, err := json.Marshal(map[string]interface{}{
			"pass":  true,
			"score": 0,
			"messages": []string{
				"hoge",
				"fuga",
			},
		})
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		fmt.Println(string(resultBytes))

		return nil
	},
}
