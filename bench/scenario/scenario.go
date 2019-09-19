package scenario

import (
	"context"
)

// Scenario isutrain のベンチマークシナリオ
type Scenario interface {
	Run(ctx context.Context) error
}
