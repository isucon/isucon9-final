package scenario

import (
	"context"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

func Initialize(ctx context.Context, baseURL string) {
	client := isutrain.NewIsutrain(baseURL)

	resp, err := client.Initialize(ctx)
	if err != nil {
		bencherror.BenchmarkErrs.AddError(err)
	}
	defer resp.Body.Close()
}
