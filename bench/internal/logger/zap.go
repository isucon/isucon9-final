package logger

import (
	"bytes"
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitZapLogger はzapロガーを初期化します
func InitZapLogger(w io.Writer) *zap.SugaredLogger {
	config := zap.NewProductionConfig()
	config.DisableCaller = true
	config.DisableStacktrace = true
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.OutputPaths = []string{"stderr"}
	config.ErrorOutputPaths = []string{"stderr"}

	var (
		enc    = zapcore.NewConsoleEncoder(config.EncoderConfig)
		syncer = zapcore.AddSync(w)
	)

	l := zap.New(zapcore.NewCore(enc, syncer, config.Level)).Named("isutrain-benchmarker")
	zap.ReplaceGlobals(l)

	return l.Sugar()
}

// InitBufferedZapLogger は内部ログバッファに書き出すzapロガーを初期化します
func InitBufferedZapLogger(buf *bytes.Buffer) *zap.SugaredLogger {
	return InitZapLogger(io.Writer(buf))
}
