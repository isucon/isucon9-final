package logger

import (
	"bytes"
	"fmt"
	"testing"

	"go.uber.org/zap"
)

func TestZapLogger(t *testing.T) {
	var logBuffer bytes.Buffer
	s := InitBufferedZapLogger(&logBuffer)
	s.Infof("hello %d", 123)

	s = zap.S()
	s.Infof("hello %d", 456)

	fmt.Println("print buffer bytes")
	fmt.Println(string(logBuffer.Bytes()))
}
