package go_lib_logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	Logger struct {
		*zap.Logger
	}
)

func NewLogger(encodeAsJSON bool, config zapcore.EncoderConfig, l zapcore.Level, output ...zapcore.WriteSyncer) *Logger {
	encoder := zapcore.NewConsoleEncoder(config)
	if encodeAsJSON {
		encoder = zapcore.NewJSONEncoder(config)
	}

	var writers []zapcore.WriteSyncer
	if len(output) == 0 {
		// set default writer
		writers = append(writers, os.Stdout)
	} else {
		writers = output
	}

	zapLogger := zap.New(zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(writers...), zap.NewAtomicLevelAt(l)))
	return &Logger{zapLogger}
}
