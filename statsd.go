package go_lib_logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"go.uber.org/zap"
)

const (
	udpSocket socketType = "udp"
	tcpSocket socketType = "tcp"
)

var (
	ErrNotConnected     = errors.New("cannot send stats, not connected to StatsD server")
	ErrEmptyUnmarshaler = errors.New("cannot parse input data, unmarshaler is not defined")

	statsdKey = []byte("\"statsd\":")

	bytesPool = sync.Pool{
		New: func() interface{} { return bytes.Buffer{} },
	}
)

type (
	unmarshalerFunc func(p []byte) (StatsdData, error)

	socketType string

	Tags map[string]string

	StatsdData struct {
		Type  string      `json:"type"`
		Name  string      `json:"name"`
		Value interface{} `json:"value"`
		Tags  Tags        `json:"tags"`
	}

	StatsdClient struct {
		conn        net.Conn
		addr        string
		prefix      string
		sockType    socketType
		timeout     time.Duration
		sampleRate  int
		unmarshaler unmarshalerFunc
	}
)

func NewStatsdClient(addr string, prefix string, timeout time.Duration, sampleRate int) *StatsdClient {
	return &StatsdClient{
		addr:        addr,
		prefix:      prefix,
		timeout:     timeout,
		sampleRate:  sampleRate,
		unmarshaler: defaultUnmarshaler,
	}
}

func (c *StatsdClient) SetUnmarshaler(f unmarshalerFunc) {
	c.unmarshaler = f
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

// Write WriteSyncer interface implementation
func (c *StatsdClient) Write(p []byte) (n int, err error) {
	n = len(p)

	if !bytes.Contains(p, statsdKey) {
		return
	}

	if c.unmarshaler == nil {
		err = ErrEmptyUnmarshaler
		n = 0
		return
	}

	data, err := c.unmarshaler(p)
	if err != nil {
		n = 0
		return
	}

	if c.sockType == tcpSocket {
		p = append(p, '\n')
	}

	sampleRate := float32(1)
	if c.sampleRate != 100 {
		sampleRate = float32(c.sampleRate) / 100
	}

	value := int64(data.Value.(float64))
	switch data.Type {
	case "gauge":
		if _, err = c.GaugeSampled(data.Name, value, data.Tags, sampleRate); err != nil {
			n = 0
		}
	case "counter":
		if _, err = c.CounterSampled(data.Name, value, data.Tags, sampleRate); err != nil {
			n = 0
		}
	case "increment":
		if _, err = c.IncrementSampled(data.Name, value, data.Tags, sampleRate); err != nil {
			n = 0
		}
	case "decrement":
		if _, err = c.DecrementSampled(data.Name, value, data.Tags, sampleRate); err != nil {
			n = 0
		}
	case "timing":
		if _, err = c.TimingSampled(data.Name, value, data.Tags, sampleRate); err != nil {
			n = 0
		}
	case "set":
		if _, err = c.Set(data.Name, value, data.Tags); err != nil {
			n = 0
		}
	case "gauge_delta":
		if _, err = c.GaugeDelta(data.Name, value, data.Tags); err != nil {
			n = 0
		}
	case "histogram":
		if _, err = c.Histogram(data.Name, value, data.Tags); err != nil {
			n = 0
		}
	}

	return
}

// Sync WriteSyncer interface implementation
func (c *StatsdClient) Sync() error {
	return nil
}

func (c *StatsdClient) FieldZap(t, name string, value interface{}, tags Tags) zap.Field {
	return zap.Any("statsd", map[string]interface{}{
		"type":  t,
		"name":  name,
		"value": value,
		"tags":  tags,
	})
}

func (c *StatsdClient) FieldZero(t, name string, value interface{}, tags Tags) *zerolog.Event {
	return zerolog.Dict().Dict("statsd",
		zerolog.Dict().
			Str("type", t).
			Str("name", name).
			Interface("value", value).
			Interface("tags", tags))
}

func (c *StatsdClient) send(stat string, format string, value interface{}, tags Tags, sampleRate float32) (int, error) {
	if c.conn == nil {
		return 0, ErrNotConnected
	}

	buff := bytesPool.Get().(bytes.Buffer)
	defer func() {
		buff.Reset()
		bytesPool.Put(buff)
	}()

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
		rNum := generator.Float32()
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

	if c.sockType == tcpSocket {
		buff.WriteByte('\n')
	}

	return fmt.Fprint(c.conn, buff.String())
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

func defaultUnmarshaler(p []byte) (StatsdData, error) {
	data := struct {
		Statsd StatsdData `json:"statsd"`
	}{}
	err := json.Unmarshal(p, &data)

	return data.Statsd, err
}
