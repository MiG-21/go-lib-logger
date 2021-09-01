package go_lib_logger_test

import (
	"testing"

	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	go_lib_logger "github.com/MiG-21/go-lib-logger"
)

type (
	NullWriter struct{}
)

func (NullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (NullWriter) Sync() error {
	return nil
}

var (
	discard = NullWriter{}
)

func BenchmarkZapLogger(b *testing.B) {
	config := zapcore.EncoderConfig{
		MessageKey: "message",
	}
	logger := go_lib_logger.NewLogger(true, config, go_lib_logger.InfoLevel, discard)

	with := zap.Any("context", map[string]interface{}{
		"field1": "field1",
		"field2": "field2",
		"field3": "field3",
		"field4": "field4",
	})

	for i := 0; i < b.N; i++ {
		logger.With(with).Log(go_lib_logger.InfoLevel, "some log message")
	}
}

func BenchmarkZeroLogger(b *testing.B) {
	logger := go_lib_logger.InitLogger(discard)
	meta := zerolog.Dict().
		Str("key1", "val1").
		Str("key2", "val2").
		Str("key3", "val3")
	with := []string{"foo", "foo1", "foo2"}

	for i := 0; i < b.N; i++ {
		logger.With(with).Logm(go_lib_logger.InfoTag, "some awesome message", meta, "foo3", "foo4", "foo5")
	}
}
