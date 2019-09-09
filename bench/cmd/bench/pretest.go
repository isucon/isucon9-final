package main

import "github.com/urfave/cli"

var preTest = cli.Command{
	Name:  "pretest",
	Usage: "PreTest実行",
	Action: func(cliCtx *cli.Context) error {
		return nil
	},
}
