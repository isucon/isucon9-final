package bencherror

import (
	"github.com/morikuni/failure"
	"go.uber.org/zap"
)

// failure.Codeを抽出
func extractCode(err error) (msg string, code failure.Code, ok bool) {
	msg, ok = failure.MessageOf(err)
	if !ok {
		return
	}

	code, ok = failure.CodeOf(err)
	if !ok {
		return
	}

	lgr := zap.S()
	if causeErr := failure.CauseOf(err); err != nil {
		lgr.Warnw("エラー発生",
			"error", err.Error(),
			"cause_error", causeErr.Error(),
		)
	}

	return
}
