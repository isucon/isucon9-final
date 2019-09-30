package main

import (
	"context"
	"errors"

	"go.uber.org/zap"

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
	lgr := zap.S()

	weight := int64(config.AvailableDays * config.WorkloadMultiplier)
	lgr.Infof("負荷レベル Lv:%d", weight)
	return &benchmarker{sem: semaphore.NewWeighted(weight)}
}

// ベンチ負荷の１単位. これの回転数を上げていく
func (b *benchmarker) load(ctx context.Context) error {
	defer b.sem.Release(1)

	scenario.NormalScenario(ctx)

	scenario.NormalCancelScenario(ctx)

	scenario.AttackReserveForOtherReservation(ctx)

	scenario.AttackReserveRaceCondition(ctx)

	scenario.AbnormalReserveWrongSection(ctx)

	scenario.AbnormalReserveWrongSeat(ctx)

	scenario.NormalManyAmbigiousSearchScenario(ctx, 5) // 負荷レベルに合わせて大きくする

	scenario.NormalManyCancelScenario(ctx, 2) // FIXME: 負荷レベルが上がってきたらあyる

	scenario.NormalVagueSearchScenario(ctx)

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
