package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const loggerName = "isutrain-benchmarker"

// InitZapLogger はzapロガーを初期化します
func InitZapLogger() (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.Encoding = "console"
	config.DisableCaller = true
	config.DisableStacktrace = true
	config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.LevelKey = ""
	config.OutputPaths = []string{"stderr"}
	config.ErrorOutputPaths = []string{"stderr"}

	l, err := config.Build()
	if err != nil {
		return nil, err
	}

	zap.ReplaceGlobals(l.Named(loggerName))

	return zap.S(), nil
}
