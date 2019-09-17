package main

import (
	"context"

	"github.com/chibiegg/isucon9-final/bench/scenario"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type benchmarker struct {
	BaseURL string

	eg *errgroup.Group
}

func newBenchmarker(baseURL string) *benchmarker {
	return &benchmarker{
		BaseURL: baseURL,
		eg:      &errgroup.Group{},
	}
}

// ベンチ負荷の１単位. これの回転数を上げていく
func (b *benchmarker) load(ctx context.Context) error {
	return nil
	// TODO: Load１単位で同期ポイント
	scenario, err := scenario.NewBasicScenario(b.BaseURL)
	if err != nil {
		return err
	}

	return scenario.Run(ctx)
}

func (b *benchmarker) run(ctx context.Context) error {
	lgr := zap.S()

	// 負荷１
	b.eg.Go(func() error {
		lgr.Info("run load 1")
		return b.load(ctx)
	})

	return b.eg.Wait()
}
