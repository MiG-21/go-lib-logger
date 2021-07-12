package go_lib_logger_test

import (
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	go_lib_logger "github.com/MiG-21/go-lib-logger"
)

func TestZapStatsd(t *testing.T) {
	addr := ":9999"
	message := "test_foo:1|g|#tag1:1\n"
	var err error

	var l net.Listener
	if l, err = net.Listen("tcp", addr); err != nil {
		t.Errorf("failed to listen: %s", err.Error())
	}
	defer func() {
		_ = l.Close()
	}()

	go func() {
		std := go_lib_logger.NewStatsdClient(addr, "test_", 1*time.Second, 100)
		if err = std.CreateTCPSocket(); err != nil {
			t.Errorf("failed to create socket: %s", err.Error())
		}
		defer func() {
			_ = std.Close()
		}()
		config := zapcore.EncoderConfig{
			TimeKey:    "T",
			MessageKey: "message",
			EncodeTime: zapcore.ISO8601TimeEncoder,
		}
		logger := go_lib_logger.NewLogger(true, config, zapcore.InfoLevel, os.Stdout, std)
		field := std.Field("gauge", "foo", 1, go_lib_logger.Tags{"tag1": "1"})
		with := zap.Any("context", map[string]interface{}{
			"field1": "field1",
			"field2": "field2",
			"field3": "field3",
			"field4": "field4",
		})
		logger.With(with).Info("some log message", field)
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer func() {
			_ = conn.Close()
		}()

		buf, err := ioutil.ReadAll(conn)
		if err != nil {
			t.Error(err)
			return
		}

		if msg := string(buf[:]); msg != message {
			t.Fatalf("Unexpected message:\nGot:\t\t%s\nExpected:\t%s\n", msg, message)
		}
		return // Done
	}
}
