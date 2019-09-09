package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/session"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/mock"
	"github.com/chibiegg/isucon9-final/bench/scenario"
	"github.com/urfave/cli"
)

var (
	paymentURI, targetURI string
	dataDir, staticDir    string
)

var (
	// デバッグフラグ
	debug bool
)

var bench = cli.Command{
	Name:  "bench",
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
		if debug {
			m := mock.Register()
			m.SetDelay(10 * time.Second)
		}

		client := isutrain.NewIsutrain("http://127.0.0.1:8000/")

		// initialize
		initSess, err := session.NewSessionForInitialize()
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		_, err = client.Initialize(context.Background(), initSess)
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		// pretest
		err = scenario.PreTest()
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		// bench
		ctx, cancel := context.WithTimeout(context.Background(), config.BenchmarkTimeout)
		defer cancel()

		benchmarker := newBenchmarker("http://127.0.0.1:8000/")
		err = benchmarker.run(ctx)
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		// posttest
		err = scenario.PostTest()
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		// 最終結果をstdoutへ書き出す
		fmt.Println(json.Marshal(map[string]interface{}{
			"pass":  true,
			"score": 0,
			"messages": []string{
				"hoge",
				"fuga",
			},
		}))

		return nil
	},
}
