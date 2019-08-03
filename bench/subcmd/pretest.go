package subcmd

import "github.com/urfave/cli"

var PreTest = cli.Command{
	Name:      "pretest",
	Usage:     "PreTest実行",
	ArgsUsage: " ",
	Action: func(cliCtx *cli.Context) error {
		return nil
	},
}
