package go_lib_logger

import (
	"math/rand"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel = zapcore.DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel = zapcore.InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel = zapcore.WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel = zapcore.ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel = zapcore.DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel = zapcore.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = zapcore.FatalLevel
)

type (
	ZapLogger struct {
		*zap.Logger
		SampleRate float32
	}
)

var (
	generator = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func NewLogger(encodeAsJSON bool, config zapcore.EncoderConfig, l zapcore.Level, output ...zapcore.WriteSyncer) *ZapLogger {
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

	return &ZapLogger{Logger: zapLogger, SampleRate: 1}
}

// With wrapper to save SampleRate option
func (log *ZapLogger) With(fields ...zap.Field) *ZapLogger {
	if len(fields) == 0 {
		return log
	}
	l := log.clone()
	l.Logger = log.Logger.With(fields...)

	return l
}

// Log wrapper to implement custom sampler logic
func (log *ZapLogger) Log(lvl zapcore.Level, msg string, fields ...zap.Field) {

	if log.SampleRate < 1 {
		rNum := generator.Float32()
		if rNum <= log.SampleRate {
			return
		}
	}

	if ce := log.Check(lvl, msg); ce != nil {
		ce.Write(fields...)
	}
}

func (log *ZapLogger) clone() *ZapLogger {
	l := *log

	return &l
}
