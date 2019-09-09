package main

import (
	"context"
	"log"

	"github.com/chibiegg/isucon9-final/bench/internal/logger"
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

func (b *benchmarker) run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("canceled")
		}
	}

	return nil
}
