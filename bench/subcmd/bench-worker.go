package subcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
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

// dequeue関連
type (
	Server struct {
		ID            int    `json:"id"`
		Hostname      string `json:"hostname"`
		GlobalIP      string `json:"global_ip"`
		PrivateIP     string `json:"private_ip"`
		IsBenchTarget bool   `json:"is_bench_target"`
	}
	Team struct {
		ID      int       `json:"id"`
		Owner   int       `json:"owner"`
		Name    string    `json:"name"`
		Servers []*Server `json:"servers"`
	}
	Job struct {
		ID     int    `json:"id"`
		Team   *Team  `json:"team"`
		Status string `json:"status"`
		Score  int    `json:"score"`
		Reason string `json:"reason"`
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
	}
)

// report関連
type (
	Result struct {
		ID       int    `json:"id"`
		Status   string `json:"status"`
		Score    int    `json:"score"`
		IsPassed bool   `json:"is_passed"`
		Reason   string `json:"reason"`
		Stdout   string `json:"stdout"`
		Stderr   string `json:"stderr`
	}
)

// Dequeue は、ポータルサーバからジョブをdequeueします
func dequeue(ctx context.Context) (*Job, error) {
	uri := fmt.Sprintf("%s/internal/job/dequeue/", portalBaseURI)

	req, err := http.NewRequest(http.MethodPost, uri, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errJobNotFound
	}

	var job *Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, err
	}

	return job, nil
}

// ベンチマーカー実行ファイルを実行
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

// Report は、ベンチマーク結果をポータルサーバに通知します
func report(ctx context.Context, jobID int, result *Result) error {
	uri := fmt.Sprintf("%s/internal/job/%d/report/", portalBaseURI, jobID)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(result); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, uri, &buf)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errReportFailed
	}

	return nil
}

var BenchWorker = cli.Command{
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
