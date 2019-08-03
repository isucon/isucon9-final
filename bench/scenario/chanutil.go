package scenario

import "context"

func SendRequestResultOrDone(ctx context.Context, resultCh chan<- *RequestResult, result *RequestResult) {
	select {
	case <-ctx.Done():
	case resultCh <- result:
	}
}
