package go_lib_logger

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"
)

var (
	ErrNotConnected = errors.New("cannot send stats, not connected to StatsD server")
)

type socketType string

type Tags map[string]string

const (
	udpSocket socketType = "udp"
	tcpSocket socketType = "tcp"
)

type StatsdClient struct {
	conn     net.Conn
	addr     string
	prefix   string
	sockType socketType
	timeout  time.Duration
}

func NewStatsdClient(addr string, prefix string, timeout time.Duration) *StatsdClient {
	return &StatsdClient{
		addr:    addr,
		prefix:  prefix,
		timeout: timeout,
	}
}

func (c *StatsdClient) CreateUDPSocket() error {
	conn, err := net.DialTimeout(string(udpSocket), c.addr, c.timeout)
	if err != nil {
		return err
	}
	c.conn = conn
	c.sockType = udpSocket
	return nil
}

func (c *StatsdClient) CreateTCPSocket() error {
	conn, err := net.DialTimeout(string(tcpSocket), c.addr, c.timeout)
	if err != nil {
		return err
	}
	c.conn = conn
	c.sockType = tcpSocket
	return nil
}

func (c *StatsdClient) Close() error {
	if nil == c.conn {
		return nil
	}
	return c.conn.Close()
}

func (c *StatsdClient) Gauge(stat string, value int64, tags Tags) (int, error) {
	return c.GaugeSampled(stat, value, tags, 1)
}

func (c *StatsdClient) GaugeSampled(stat string, value int64, tags Tags, sampleRate float32) (int, error) {
	return c.send(stat, "%d|g", value, tags, sampleRate)
}

func (c *StatsdClient) GaugeDelta(stat string, value int64, tags Tags) (int, error) {
	if value < 0 {
		return c.send(stat, "-%d|g", value, tags, 1)
	}
	return c.send(stat, "+%d|g", value, tags, 1)
}

func (c *StatsdClient) Set(stat string, count int64, tags Tags) (int, error) {
	return c.send(stat, "%d|s", count, tags, 1)
}

func (c *StatsdClient) Counter(stat string, count int64, tags Tags) (int, error) {
	return c.CounterSampled(stat, count, tags, 1)
}

func (c *StatsdClient) CounterSampled(stat string, count int64, tags Tags, sampleRate float32) (int, error) {
	return c.send(stat, "%d|c", count, tags, sampleRate)
}

func (c *StatsdClient) Increment(stat string, count int64, tags Tags) (int, error) {
	return c.Counter(stat, count, tags)
}

func (c *StatsdClient) IncrementSampled(stat string, count int64, tags Tags, sampleRate float32) (int, error) {
	return c.CounterSampled(stat, count, tags, sampleRate)
}

func (c *StatsdClient) Decrement(stat string, count int64, tags Tags) (int, error) {
	return c.Counter(stat, -count, tags)
}

func (c *StatsdClient) DecrementSampled(stat string, count int64, tags Tags, sampleRate float32) (int, error) {
	return c.CounterSampled(stat, -count, tags, sampleRate)
}

func (c *StatsdClient) Timing(stat string, delta int64, tags Tags) (int, error) {
	return c.TimingSampled(stat, delta, tags, 1)
}

func (c *StatsdClient) TimingSampled(stat string, delta int64, tags Tags, sampleRate float32) (int, error) {
	return c.send(stat, "%d|ms", delta, tags, sampleRate)
}

func (c *StatsdClient) Histogram(stat string, delta int64, tags Tags) (int, error) {
	return c.send(stat, "%d|h", delta, tags, 1)
}

func (c *StatsdClient) Raw(metricString string) (int, error) {
	if c.sockType == tcpSocket {
		metricString += "\n"
	}

	return fmt.Fprint(c.conn, metricString)
}

func (c *StatsdClient) send(stat string, format string, value interface{}, tags Tags, sampleRate float32) (int, error) {
	if c.conn == nil {
		return 0, ErrNotConnected
	}

	buff := bytes.Buffer{}
	if _, err := buff.WriteString(c.prefix); err != nil {
		return 0, err
	}
	if _, err := buff.WriteString(stat); err != nil {
		return 0, err
	}
	if _, err := buff.WriteString(":"); err != nil {
		return 0, err
	}
	if _, err := buff.WriteString(fmt.Sprintf(format, value)); err != nil {
		return 0, err
	}

	if sampleRate < 1 {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		rNum := r.Float32()
		if rNum <= sampleRate {
			if _, err := buff.WriteString(fmt.Sprintf("|@%f", sampleRate)); err != nil {
				return 0, err
			}
		} else {
			return 0, nil
		}
	}

	if _, err := formatTags(tags, &buff); err != nil {
		return 0, err
	}

	return c.Raw(buff.String())
}

func formatTags(tags Tags, buff *bytes.Buffer) (int, error) {
	var b int
	if len(tags) > 0 {
		if bb, err := buff.WriteString("|#"); err != nil {
			return 0, err
		} else {
			b += bb
		}

		ln := len(tags)
		i := 0
		for k, v := range tags {
			if bb, err := buff.WriteString(k); err != nil {
				return 0, err
			} else {
				b += bb
			}
			if bb, err := buff.WriteString(":"); err != nil {
				return 0, err
			} else {
				b += bb
			}
			if bb, err := buff.WriteString(v); err != nil {
				return 0, err
			} else {
				b += bb
			}
			i += 1
			if i < ln {
				if bb, err := buff.WriteString(","); err != nil {
					return 0, err
				} else {
					b += bb
				}
			}
		}
	}
	return b, nil
}
