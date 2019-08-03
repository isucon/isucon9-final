package util

import (
	"context"

	"github.com/chibiegg/isucon9-final/bench/internal/result"
)

func DoneOrSendBenchResult(ctx context.Context, resultCh chan<- *result.BenchResult, result *result.BenchResult) {
	select {
	case <-ctx.Done():
	case resultCh <- result:
	}
}
