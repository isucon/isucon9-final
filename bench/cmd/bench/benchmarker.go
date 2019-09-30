package main

import (
	"context"
	"errors"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/scenario"
	"golang.org/x/sync/semaphore"
)

var (
	ErrBenchmarkFailure = errors.New("ベンチマークに失敗しました")
)

type benchmarker struct {
	sem *semaphore.Weighted
}

func newBenchmarker() *benchmarker {
	return &benchmarker{sem: semaphore.NewWeighted(int64(config.AvailableDays * config.WorkloadMultiplier))}
}

// ベンチ負荷の１単位. これの回転数を上げていく
func (b *benchmarker) load(ctx context.Context) error {
	defer b.sem.Release(1)

	scenario.NormalScenario(ctx)

	scenario.NormalCancelScenario(ctx)

	scenario.AttackReserveForOtherReservation(ctx)

	// FIXME: webappの課金情報がおかしくなる
	// scenario.AttackReserveRaceCondition(ctx)

	scenario.AbnormalReserveWrongSection(ctx)

	scenario.AbnormalReserveWrongSeat(ctx)

	scenario.NormalManyAmbigiousSearchScenario(ctx, 5) // 負荷レベルに合わせて大きくする

	scenario.NormalManyCancelScenario(ctx, 2) // FIXME: 負荷レベルが上がってきたらあyる

	// FIXME: webappの課金情報がおかしくなる
	// scenario.NormalVagueSearchScenario(ctx)

	if config.AvailableDays > 200 { // FIXME: 値が適当
		scenario.GoldenWeekScenario(ctx)
	}

	return nil
}

func (b *benchmarker) run(ctx context.Context) error {
	defer bencherror.BenchmarkErrs.DumpCounters()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if bencherror.BenchmarkErrs.IsFailure() {
				// 失格と分かれば、早々にベンチマークを終了
				return ErrBenchmarkFailure
			}

			for i := 0; i < 5; i++ {
				if err := b.sem.Acquire(ctx, 1); err != nil {
					return ErrBenchmarkFailure
				}
				go b.load(ctx)
			}
		}
	}
}
