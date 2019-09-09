package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestZapLogger(t *testing.T) {
	s, err := InitZapLogger()
	assert.NoError(t, err)
	s.Infof("hello %d", 123)

	s = zap.S()
	s.Infof("hello %d", 456)
	s.Warnf("warn %d", 123)
}
