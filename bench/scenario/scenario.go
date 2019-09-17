package scenario

import (
	"context"
)

// Scenario isutrain のベンチマークシナリオ
type Scenario interface {
	Run(ctx context.Context) error
}

// カウンタを設けて、ここでシナリオを生成する
// func NewScenario(tick uint64, baseURL string) Scenario {
// return NewBasicScenario(baseURL)
// }
