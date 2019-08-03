package bencherror

import (
	"github.com/morikuni/failure"
)

var (
	errCritical    failure.StringCode = "error"
	errApplication failure.StringCode = "error application"
	errTimeout     failure.StringCode = "error timeout"
	errTemporary   failure.StringCode = "error temporary"
)

func NewCriticalError(err error) error {
	return failure.Translate(err, errCritical)
}

func NewApplicationError(err error) error {
	return failure.Translate(err, errApplication)
}

func NewTimeoutError(err error) error {
	return failure.Translate(err, errTimeout)
}

func NewTemporaryError(err error) error {
	return failure.Translate(err, errTemporary)
}
