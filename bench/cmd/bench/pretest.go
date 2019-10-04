package main

import (
	"context"

	"github.com/chibiegg/isucon9-final/bench/assets"
	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/logger"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/payment"
	"github.com/chibiegg/isucon9-final/bench/scenario"
	"github.com/urfave/cli"
)

var pretest = cli.Command{
	Name:  "pretest",
	Usage: "Pretest実施",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "payment",
			Value:       "http://localhost:5000",
			Destination: &config.PaymentBaseURL,
			EnvVar:      "BENCH_PAYMENT_URL",
		},
		cli.StringFlag{
			Name:        "target",
			Value:       "http://localhost",
			Destination: &config.TargetBaseURL,
			EnvVar:      "BENCH_TARGET_URL",
		},
		cli.StringFlag{
			Name:        "assetdir",
			Value:       "assets/testdata",
			Destination: &assetDir,
			EnvVar:      "BENCH_ASSETDIR",
		},
	},
	Action: func(cliCtx *cli.Context) error {
		ctx := context.Background()

		lgr, err := logger.InitZapLogger()
		if err != nil {
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 1)
		}

		lgr.Info("===== Prepare benchmarker =====")

		assets, err := assets.Load(assetDir)
		if err != nil {
			lgr.Warn("静的ファイルをローカルから読み出せませんでした: %+v", err)
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 1)
		}

		initClient, err := isutrain.NewClientForInitialize()
		if err != nil {
			lgr.Warn("isutrainクライアント生成に失敗しました: %+v", err)
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 1)
		}

		testClient, err := isutrain.NewClient()
		if err != nil {
			lgr.Warn("pretestクライアント生成に失敗しました: %+v", err)
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 1)
		}

		paymentClient, err := payment.NewClient()
		if err != nil {
			lgr.Warn("課金クライアント生成に失敗しました: %+v", err)
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 1)
		}

		lgr.Info("===== Initialize webapp =====")
		initClient.Initialize(ctx)
		if bencherror.InitializeErrs.IsError() {
			lgr.Warnf("webappへの /initialize でエラーが発生: %+v", bencherror.InitializeErrs.Msgs)
			dumpFailedResult(bencherror.InitializeErrs.Msgs)
			return nil
		}

		// pretest (まず、正しく動作できているかチェック. エラーが見つかったら、採点しようがないのでFAILにする)
		lgr.Info("===== Pretest webapp =====")
		scenario.Pretest(ctx, testClient, paymentClient, assets)
		if bencherror.PreTestErrs.IsError() {
			lgr.Warnf("webappへの pretest でエラーが発生: %+v", bencherror.PreTestErrs.Msgs)
			dumpFailedResult(bencherror.PreTestErrs.Msgs)
			return nil
		}

		lgr.Info("Passed.")

		return nil
	},
}
