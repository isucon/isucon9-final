package main

import (
	"context"
	"log"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"github.com/chibiegg/isucon9-final/bench/internal/result"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

type Benchmarker struct {
	BaseURL  string
	resultCh chan *result.BenchResult

	eg *errgroup.Group

	additionalScenarios []isutrain.Scenario
	// tick はシナリオ生成の際、どのシナリオを用いるか決定する際の判断材料として用います
	tick uint64
}

func NewBenchmarker(baseURL string) *Benchmarker {
	return &Benchmarker{
		BaseURL:  baseURL,
		resultCh: make(chan *result.BenchResult, 2000),
	}
}

// PreTest は、ベンチマーク実行前の整合性チェックを実施します
func (b *Benchmarker) PreTest() {

}

// Bench はベンチマークを実施します
func (b *Benchmarker) Bench(ctx context.Context) error {
	atomic.AddUint64(&b.tick, 1)
	scenario := isutrain.NewScenario(b.tick, b.BaseURL)

	err := scenario.Run(ctx)
	if err != nil {
		return err
	}

	var sum int
	for benchResult := range scenario.BenchResultStream() {
		sum += benchResult.Type.Score()
		if benchResult.Err != nil {
			log.Printf("error = %+v\n", benchResult.Err)
		}
	}

	log.Printf("score = %d\n", sum)

	return nil
}
