package main

import (
	"encoding/json"
	"fmt"
	"log"

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
		// if debug {
		// 	m := mock.Register()
		// 	m.SetDelay(10 * time.Second)
		// }

		// client := isutrain.NewIsutrain("http://127.0.0.1:8000/")

		// // initialize
		// initSess, err := session.NewSessionForInitialize()
		// if err != nil {
		// 	return cli.NewExitError(err, 1)
		// }
		// _, err = client.Initialize(context.Background(), initSess)
		// if err != nil {
		// 	return cli.NewExitError(err, 1)
		// }

		// // pretest
		// err = scenario.PreTest()
		// if err != nil {
		// 	return cli.NewExitError(err, 1)
		// }

		// // bench
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

		// // posttest
		// err = scenario.PostTest()
		// if err != nil {
		// 	return cli.NewExitError(err, 1)
		// }

		log.Printf("payment   = %s\n", paymentURI)
		log.Printf("target    = %s\n", targetURI)
		log.Printf("datadir   = %s\n", dataDir)
		log.Printf("staticdir = %s\n", staticDir)

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
