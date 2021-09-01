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

	go_lib_logger "github.com/MiG-21/go-lib-logger"
	"github.com/rs/zerolog"
)

func TestZerologStatsd(t *testing.T) {
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
		std.SetUnmarshaler(go_lib_logger.ZeroStatsdUnmarshaler)
		logger := go_lib_logger.InitLogger(os.Stdout, std)
		field := std.FieldZero("gauge", "foo", 1, go_lib_logger.Tags{"tag1": "1"})
		logger.With([]string{"foo", "foo1", "foo2"}).Logm(go_lib_logger.InfoTag, "some awesome message", field, "foo3", "foo4", "foo5")
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

func TestZerologTagsStdout(t *testing.T) {
	message := `{"level":"info","tags":["level:info","foo","foo1","foo2","foo3","foo4","foo5"],"message":"some awesome message"}`
	// keep backup of the real stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := go_lib_logger.InitLogger(os.Stdout)
	logger.With([]string{"foo", "foo1", "foo2"}).Log(go_lib_logger.InfoTag, "some awesome message", "foo3", "foo4", "foo5")

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

func TestZerologMetaTagsStdout(t *testing.T) {
	message := `{"level":"info","meta":{"key1":"val1","key2":"val2","key3":"val3"},"tags":["level:info","foo","foo1","foo2","foo3","foo4","foo5"],"message":"some awesome message"}`
	// keep backup of the real stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := go_lib_logger.InitLogger(os.Stdout)
	meta := zerolog.Dict().
		Str("key1", "val1").
		Str("key2", "val2").
		Str("key3", "val3")
	logger.With([]string{"foo", "foo1", "foo2"}).Logm(go_lib_logger.InfoTag, "some awesome message", meta, "foo3", "foo4", "foo5")

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
