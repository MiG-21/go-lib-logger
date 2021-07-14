# go-lib-logger

## Usage

### Initialization
```go
// addr - statsd server address
// "test_" - prefix
// 1*time.Second - timeout
// 100 - sample rate
std := go_lib_logger.NewStatsdClient(addr, "test_", 1*time.Second, 100)
```

### TCP connection
```go
if err = std.CreateTCPSocket(); err != nil {
    // some error handling
}
defer func() {
_ = std.Close()
}()
```

### UDP connection
```go
if err = std.CreateUDPSocket(); err != nil {
    // some error handling
}
defer func() {
_ = std.Close()
}()
```

### send metrics
```go
if _, err = std.Gauge("foo", 1, go_lib_logger.Tags{"tag1": "1"}); err != nil {
    // some error handling
}
```
### metrics methods
* Gauge(stat string, value int64, tags Tags) (int, error)
* GaugeSampled(stat string, value int64, tags Tags, sampleRate float32) (int, error)
* GaugeDelta(stat string, value int64, tags Tags) (int, error)
* Set(stat string, count int64, tags Tags) (int, error)
* Counter(stat string, count int64, tags Tags) (int, error)
* CounterSampled(stat string, count int64, tags Tags, sampleRate float32) (int, error)
* Increment(stat string, count int64, tags Tags) (int, error)
* IncrementSampled(stat string, count int64, tags Tags, sampleRate float32) (int, error)
* Decrement(stat string, count int64, tags Tags) (int, error)
* DecrementSampled(stat string, count int64, tags Tags, sampleRate float32) (int, error)
* Timing(stat string, delta int64, tags Tags) (int, error)
* TimingSampled(stat string, delta int64, tags Tags, sampleRate float32) (int, error)
* Histogram(stat string, delta int64, tags Tags) (int, error)
