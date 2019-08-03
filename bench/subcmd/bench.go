package subcmd

const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
	StatusTimeout = "timeout"
)

// execBench(ベンチマーカーを外部コマンド実行)関連
type (
	BenchmarkerParams struct {
		PaymentURI string
		TargetURI  string
		DataDir    string
		StaticDir  string
	}
)

func runBenchmarker(cliCtx *cli.Context) error {
	return nil
}