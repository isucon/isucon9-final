package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/urfave/cli"
)

var (
	portalBaseURI, paymentBaseURI, targetBaseURI string
	benchmarkerPath                              string

	dequeueInterval int

	retryLimit, retryInterval int
)

var (
	errJobNotFound      = errors.New("ジョブが見つかりませんでした")
	errReportFailed     = errors.New("ベンチ結果報告に失敗しました")
	errAllowIPsNotFound = errors.New("許可すべきIPが見つかりませんでした")
)

const (
	MsgTimeout = "ベンチマーク処理がタイムアウトしました"
	MsgFail    = "運営に連絡してください"
)

const (
	StatusSuccess = "done"
	StatusFailed  = "fail"
	StatusTimeout = "timeout"
)

// ベンチマーカー実行ファイルを実行
// FIXME: リトライ
func execBench(ctx context.Context, job *Job) (*Result, error) {
	// ターゲットサーバを取得
	targetServer, err := getTargetServer(job)
	if err != nil {
		return nil, err
	}

	// 許可IP一覧を取得
	allowIPs := getAllowIPs(job)
	if len(allowIPs) == 0 {
		return nil, errAllowIPsNotFound
	}

	var (
		paymentUri = "https://localhost:5000"
		targetUri  = fmt.Sprintf("http://%s", targetServer.GlobalIP)

		dataDir   = "/home/isucon/isutrain/initial-data"
		staticDir = "/home/isucon/isutrain/webapp/public/static"
	)

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, benchmarkerPath, []string{
		"run",
		"--payment=" + paymentUri,
		"--target=" + targetUri,
		"--datadir=" + dataDir,
		"--staticdir=" + staticDir,
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

	// ベンチ結果をUnmarshal
	var result *BenchResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		log.Printf("json unmarshal error: %+v\n", err)
		log.Println(string(stdout.Bytes()))
		return nil, err
	}

	// FIXME: result.Messagesの扱い
	return &Result{
		Stdout:   string(stdout.Bytes()),
		Stderr:   string(stderr.Bytes()),
		Reason:   strings.Join(result.Messages, "\n"),
		IsPassed: result.Pass,
		Score:    result.Score,
		Status:   status,
	}, nil
}

var run = cli.Command{
	Name:  "benchworker",
	Usage: "ベンチマークワーカー実行",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "portal",
			Value:       "http://localhost:8000",
			Destination: &portalBaseURI,
		},
		cli.StringFlag{
			Name:        "benchmarker",
			Value:       "/home/isucon/isutrain/bin/benchmarker",
			Destination: &benchmarkerPath,
		},
		cli.IntFlag{
			Name:        "retrylimit",
			Value:       10,
			Destination: &retryLimit,
		},
		cli.IntFlag{
			Name:        "retryinterval",
			Value:       2,
			Destination: &retryInterval,
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
					// log.Println(err)
					continue
				}

				result, err := execBench(ctx, job)
				if err != nil {
					// FIXME: ベンチ失敗した時のaction
					break
				}

				// ポータルに結果を報告
				reportRetrier := retrier.New(retrier.ConstantBackoff(retryLimit, time.Duration(retryInterval)*time.Second), nil)
				err = reportRetrier.RunCtx(ctx, func(ctx context.Context) error {
					return report(ctx, job.ID, result)
				})
				if err != nil {
					log.Println(err)
				}
			}
		}

		return nil
	},
}
