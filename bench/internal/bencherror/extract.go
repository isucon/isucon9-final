package bencherror

import "github.com/morikuni/failure"

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

	return
}
