package main

import (
	"log"
	"os"
	"time"

	"github.com/urfave/cli"
)

func init() {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatalln(err)
	}
	time.Local = loc
}

func main() {
	os.Exit(cliMain())
}

func cliMain() int {
	app := cli.NewApp()
	app.Name = "isutrainbench"
	app.Usage = "isutrain ベンチマーカー"
	app.Description = "isutrainのベンチマークを行います"
	app.HelpName = "isutrainbench"

	app.Commands = []cli.Command{
		bench,
		preTest,
	}

	app.Action = func(cliCtx *cli.Context) error {
		return cli.ShowAppHelp(cliCtx)
	}

	if err := app.Run(os.Args); err != nil {
		exitErr := err.(*cli.ExitError)
		log.Println(exitErr.Error())
		return exitErr.ExitCode()
	}

	return 0
}
