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
	targetPort                    int
	portalBaseURI, paymentBaseURI string
	benchmarkerPath               string
	assetDir                      string

	dequeueInterval int

	retryLimit, retryInterval int

	messageLimit int
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
	StatusFailed  = "aborted"
	StatusTimeout = "aborted"
)

func joinN(messages []string, n int) string {
	if len(messages) > n {
		return strings.Join(messages[:n], ",\n")
	}
	return strings.Join(messages, ",\n")
}

// ベンチマーカー実行ファイルを実行
// FIXME: リトライ
func execBench(ctx context.Context, job *Job) (*Result, error) {
	// ターゲットサーバを取得
	targetServer, err := getTargetServer(job)
	if err != nil {
		return nil, err
	}

	targetURI := fmt.Sprintf("http://%s:%d", targetServer.GlobalIP, targetPort)

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, benchmarkerPath, []string{
		"run",
		"--payment=" + paymentBaseURI,
		"--target=" + targetURI,
		"--assetdir=" + assetDir,
	}...)
	log.Printf("exec_path=%s", cmd.Path)
	for _, arg := range cmd.Args {
		log.Printf("\t- args=%s\n", arg)
	}
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
		log.Println(string(stdout.Bytes()))
		return nil, err
	}

	// FIXME: result.Messagesの扱い
	return &Result{
		ID:       job.ID,
		Stdout:   string(stdout.Bytes()),
		Stderr:   string(stderr.Bytes()),
		Reason:   joinN(result.Messages, messageLimit),
		IsPassed: result.Pass,
		Score:    result.Score,
		Status:   status,
	}, nil
}

var run = cli.Command{
	Name:  "run",
	Usage: "ベンチマークワーカー実行",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "portal",
			Value:       "http://localhost:8000",
			Destination: &portalBaseURI,
			EnvVar:      "BENCHWORKER_PORTAL_URL",
		},
		cli.StringFlag{
			Name:        "payment",
			Value:       "http://localhost:5000",
			Destination: &paymentBaseURI,
			EnvVar:      "BENCHWORKER_PAYMENT_URL",
		},
		cli.IntFlag{
			Name:        "target-port",
			Value:       80,
			Destination: &targetPort,
			EnvVar:      "BENCHWORKER_TARGET_PORT",
		},
		cli.StringFlag{
			Name:        "assetdir",
			Value:       "/home/isucon/isutrain/assets",
			Destination: &assetDir,
			EnvVar:      "BENCHWORKER_ASSETDIR",
		},
		cli.StringFlag{
			Name:        "benchmarker",
			Value:       "/home/isucon/isutrain/bin/benchmarker",
			Destination: &benchmarkerPath,
			EnvVar:      "BENCHWORKER_BENCHMARKER_BINPATH",
		},
		cli.IntFlag{
			Name:        "retrylimit",
			Value:       10,
			Destination: &retryLimit,
			EnvVar:      "BENCHWORKER_RETRY_LIMIT",
		},
		cli.IntFlag{
			Name:        "retryinterval",
			Value:       2,
			Destination: &retryInterval,
			EnvVar:      "BENCHWORKER_RETRY_INTERVAL",
		},
		cli.IntFlag{
			Name:        "message-limit",
			Value:       10,
			Destination: &messageLimit,
			EnvVar:      "BENCHWORKER_MESSAGE_LIMIT",
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
					continue
				}
				log.Printf("dequeue job id=%d team_id=%d target_server=%+v", job.ID, job.Team.ID, job.Team.Servers)

				reportRetrier := retrier.New(retrier.ConstantBackoff(retryLimit, time.Duration(retryInterval)*time.Second), nil)

				result, err := execBench(ctx, job)
				if err != nil {
					// FIXME: ベンチ失敗した時のaction
					reportErr := reportRetrier.RunCtx(ctx, func(ctx context.Context) error {
						return report(ctx, job.ID, &Result{
							ID:       job.ID,
							Status:   StatusFailed,
							IsPassed: false,
							Reason:   err.Error(),
						})
					})
					log.Println(reportErr)
					break
				}

				// ポータルに結果を報告
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
