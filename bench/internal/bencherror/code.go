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

func NewCriticalError(err error, msg string, args ...interface{}) error {
	return failure.Translate(err, errCritical, failure.Messagef(msg, args...))
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

func NewHTTPStatusCodeError(resp *http.Response, wantStatusCode int) error {
	// prefix := fmt.Sprintf("%s %s", resp.Request.Method, resp.Request.URL.Path)
	prefix := ""
	// prefix := resp.Request.Method
	// prefix += " "
	// prefix += resp.Request.URL.String()

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
