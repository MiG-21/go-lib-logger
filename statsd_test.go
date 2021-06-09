package go_lib_logger_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"testing"
	"time"

	go_lib_logger "github.com/MiG-21/go-lib-logger"
)

func TestStatsd(t *testing.T) {
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
		std := go_lib_logger.NewStatsdClient(addr, "test_", 1 * time.Second)
		if err = std.CreateTCPSocket(); err != nil {
			t.Errorf("failed to create socket: %s", err.Error())
		}
		defer func() {
			_ = std.Close()
		}()
		if _, err = std.Gauge("foo", 1, go_lib_logger.Tags{"tag1": "1"}); err != nil {
			t.Errorf("failed to write to socket: %s", err.Error())
		}
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

		fmt.Println(string(buf[:]))
		if msg := string(buf[:]); msg != message {
			t.Fatalf("Unexpected message:\nGot:\t\t%s\nExpected:\t%s\n", msg, message)
		}
		return // Done
	}
}
