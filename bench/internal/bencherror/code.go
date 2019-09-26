package bencherror

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/morikuni/failure"
)

var (
	errCritical    failure.StringCode = "error"
	errApplication failure.StringCode = "error application"
	errTimeout     failure.StringCode = "error timeout"
	errTemporary   failure.StringCode = "error temporary"
)

func NewSimpleCriticalError(msg string, args ...interface{}) error {
	return failure.New(errCritical, failure.Messagef(msg, args...))
}

func NewCriticalError(err error, msg string, args ...interface{}) error {
	return failure.Translate(err, errCritical, failure.Messagef(msg, args...))
}

func NewSimpleApplicationError(msg string, args ...interface{}) error {
	return failure.New(errApplication, failure.Messagef(msg, args...))
}

func NewApplicationError(err error, msg string, args ...interface{}) error {
	return failure.Translate(err, errApplication, failure.Messagef(msg, args...))
}

func NewTimeoutError(err error, msg string, args ...interface{}) error {
	return failure.Translate(err, errTimeout, failure.Messagef(msg, args...))
}

func NewTemporaryError(err error, msg string, args ...interface{}) error {
	return failure.Translate(err, errTemporary, failure.Messagef(msg, args...))
}

func NewHTTPStatusCodeError(req *http.Request, resp *http.Response, wantStatusCode int) error {
	prefix := fmt.Sprintf("%s %s: ", req.Method, req.URL.Path)

	if resp.StatusCode != wantStatusCode {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return failure.Wrap(err, failure.Message(prefix+"HTTPレスポンスボディの読み込みに失敗しました; "))
		}
		return failure.Translate(
			fmt.Errorf("status_code: %d; body: %s", resp.StatusCode, body),
			errApplication,
			failure.Messagef("%sgot response status code %d; want %d", prefix, resp.StatusCode, wantStatusCode),
		)
	}
	return nil
}
