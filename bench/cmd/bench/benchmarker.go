package main

import (
	"context"
	"errors"
	"sync"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/scenario"
	"go.uber.org/zap"
)

var (
	ErrBenchmarkFailure = errors.New("ベンチマークに失敗しました")
)

type benchmarker struct {
	level int
}

// ベンチ負荷の１単位. これの回転数を上げていく
func (b *benchmarker) load(ctx context.Context) error {
	// TODO: Load１単位で同期ポイント

	scenario.NormalScenario(ctx)

	scenario.NormalCancelScenario(ctx)

	scenario.NormalAmbigiousSearchScenario(ctx)

	return nil
}

func (b *benchmarker) run(ctx context.Context) error {
	defer bencherror.BenchmarkErrs.DumpCounters()
	lgr := zap.S()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if bencherror.BenchmarkErrs.IsFailure() {
				// 失格と分かれば、早々にベンチマークを終了
				return ErrBenchmarkFailure
			}
			var wg sync.WaitGroup
			for i := 0; i < b.level; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					b.load(ctx)
				}()
			}

			wg.Wait()
			b.level++
			lgr.Infof("負荷レベルが上がります: Lv. %d", b.level)
		}
	}
}
