package main

import (
	"context"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/logger"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/urfave/cli"
)

var bgtest = cli.Command{
	Name:  "bgtest",
	Usage: "Bgtest実施",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "target",
			Value:       "http://localhost",
			Destination: &config.TargetBaseURL,
			EnvVar:      "BENCH_TARGET_URL",
		},
	},
	Action: func(cliCtx *cli.Context) error {
		ctx := context.Background()

		lgr, err := logger.InitZapLogger()
		if err != nil {
			dumpFailedResult([]string{})
			return cli.NewExitError(err, 1)
		}

		lgr.Info("===== Prepare bgtester =====")

		initClient, err := isutrain.NewClientForInitialize()
		if err != nil {
			lgr.Warn("isutrainクライアント生成に失敗しました: %+v", err)
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

		tester, err := newBgTester()
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		if err := tester.run(ctx); err != nil {
			return cli.NewExitError(err, 1)
		}

		lgr.Info("Passed.")

		return nil
	},
}
