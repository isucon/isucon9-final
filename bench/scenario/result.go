package scenario

import "github.com/chibiegg/isucon9-final/bench/internal/score"

// RequestResult はウェブアプリへのリクエスト結果です
type RequestResult struct {
	Type score.Type
	Err  error
}

func NewRequestResult(typ score.Type, err error) *RequestResult {
	return &RequestResult{
		Type: typ,
		Err:  err,
	}
}
