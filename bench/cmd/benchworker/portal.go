package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
		Stderr   string `json:"stderr"`
	}
	BenchResult struct {
		Pass     bool     `json:"pass"`
		Score    int      `json:"score"`
		Messages []string `json:"messages"`
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

// Report は、ベンチマーク結果をポータルサーバに通知します
// FIXME: リトライ
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

	req.Header.Set("Content-Type", "application/json")

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
