package main

import (
	"context"

	"github.com/chibiegg/isucon9-final/bench/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type benchmarker struct {
	BaseURL string

	eg *errgroup.Group
}

func newBenchmarker(baseURL string) (*benchmarker, error) {
	benchmarker := &benchmarker{
		BaseURL: baseURL,
	}

	_, err := logger.InitZapLogger()
	if err != nil {
		return nil, err
	}

	return benchmarker, nil
}

// ベンチ負荷の１単位. これの回転数を上げていく
func (b *benchmarker) load(ctx context.Context) error {

	return nil
}

func (b *benchmarker) run(ctx context.Context) error {
	lgr := zap.S()

	// 負荷１
	b.eg.Go(func() error {
		lgr.Warn("run load 1")
		return b.load(ctx)
	})

	return b.eg.Wait()
}
