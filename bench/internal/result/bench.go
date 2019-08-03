package result

import "github.com/chibiegg/isucon9-final/bench/internal/score"

// BenchResult はベンチマーク結果です
type BenchResult struct {
	Type score.Type
	Err  error
}

func NewBenchResult(typ score.Type, err error) *BenchResult {
	return &BenchResult{
		Type: typ,
		Err:  err,
	}
}
