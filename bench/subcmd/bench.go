package subcmd

import (
	"github.com/urfave/cli"
)

var Bench = cli.Command{
	Name:      "bench",
	Aliases:   []string{"b"},
	Usage:     "ベンチマーク実行",
	ArgsUsage: " ",
	Action: func(cliCtx *cli.Context) error {
		return nil
	},
}
