package go_lib_logger_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
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
		logger := go_lib_logger.NewLogger(true, config, go_lib_logger.InfoLevel, os.Stdout, std)
		field := std.FieldZap("gauge", "foo", 1, go_lib_logger.Tags{"tag1": "1"})
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

func TestZapStdout(t *testing.T) {
	message := `{"message":"some log message","context":{"field1":"field1","field2":"field2","field3":"field3","field4":"field4"}}`
	// keep backup of the real stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	config := zapcore.EncoderConfig{
		MessageKey: "message",
	}
	logger := go_lib_logger.NewLogger(true, config, go_lib_logger.InfoLevel, os.Stdout)
	with := zap.Any("context", map[string]interface{}{
		"field1": "field1",
		"field2": "field2",
		"field3": "field3",
		"field4": "field4",
	})
	logger.With(with).Log(go_lib_logger.InfoLevel, "some log message")

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	_ = w.Close()
	// restoring the real stdout
	os.Stdout = old
	out := <-outC

	if strings.TrimSpace(out) != message {
		t.Fatalf("Unexpected message:\nGot:\t\t%s\nExpected:\t%s\n", out, message)
	}
}
