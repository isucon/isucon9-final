package scenario

import (
	"context"
	"errors"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/session"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

var (
	ErrInitialize = errors.New("/initialize に失敗しました")
	ErrSession    = errors.New("セッション作成に失敗しました")
)

func Initialize(ctx context.Context, baseURL string) {
	sess, err := session.NewSessionForInitialize()
	if err != nil {
		bencherror.PreTestErrs.AddError(ErrSession)
	}

	client := isutrain.NewIsutrain(baseURL)

	resp, err := client.Initialize(ctx, sess)
	if err != nil {
		bencherror.PreTestErrs.AddError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bencherror.PreTestErrs.AddError(ErrInitialize)
	}
}
