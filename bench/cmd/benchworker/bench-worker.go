package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli"
)

var (
	portalBaseURI, paymentBaseURI, targetBaseURI string
	benchmarkerPath                              string

	dequeueDuration int
)

var (
	errJobNotFound  = errors.New("ジョブが見つかりませんでした")
	errReportFailed = errors.New("ベンチ結果報告に失敗しました")
)

const (
	MsgTimeout = "ベンチマーク処理がタイムアウトしました"
	MsgFail    = "運営に連絡してください"
)

const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
	StatusTimeout = "timeout"
)

// ベンチマーカー実行ファイルを実行
// FIXME: リトライ
func execBench(ctx context.Context, job *Job) (*Result, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, benchmarkerPath, []string{
		"--paymenturi=",
	}...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	status := StatusSuccess

	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Run()
	}()

	select {
	case err := <-errCh:
		if err != nil {
			status = StatusFailed
		}
	case <-ctx.Done():
		status = StatusTimeout
	}

	return &Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Status: status,
	}, nil
}

var benchWorker = cli.Command{
	Name:  "benchworker",
	Usage: "ベンチマークワーカー実行",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "benchdir",
			Value:       "/home/isucon/isutrain/bin/benchmarker",
			Destination: &benchmarkerPath,
		},
	},
	Action: func(cliCtx *cli.Context) error {
		ctx := context.Background()

		sigCh := make(chan os.Signal, 1)
		defer close(sigCh)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
	loop:
		for {
			select {
			case <-sigCh:
				break loop
			case <-ticker.C:
				job, err := dequeue(ctx)
				if err != nil {
					// dequeueが失敗しても終了しない
					log.Println(err)
				}

				result, err := execBench(ctx, job)
				if err != nil {
					// FIXME: ベンチ失敗した時のaction
					break
				}

				// FIXME: リトライ
				err = report(ctx, job.ID, result)
				if err != nil {
					// FIXME: 報告失敗した場合、Slackに投げておいたほうがいい？
				}
			}
		}

		return nil
	},
}
