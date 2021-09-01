package go_lib_logger

import (
	"encoding/json"
	"io"
	"os"

	"github.com/rs/zerolog"
)

const (
	DebugTag string = "level:debug"
	InfoTag  string = "level:info"
	WarnTag  string = "level:warn"
	FatalTag string = "level:fatal"
	PanicTag string = "level:panic"
	ErrorTag string = "level:error"
)

type CfgLogger struct {
	ImportantTags   []string
	BlacklistedTags []string
	MetaField       string
	TagsField       string
}

type ZeroLogger struct {
	zerolog.Logger
	parent     *ZeroLogger
	Tags       []string
	cfg        CfgLogger
	SampleRate float32
}

func InitLogger(writers ...io.Writer) *ZeroLogger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}
	multi := zerolog.MultiLevelWriter(writers...)

	l := &ZeroLogger{
		Logger:     zerolog.New(multi),
		SampleRate: 1,
	}
	l.cfg = CfgLogger{
		BlacklistedTags: []string{},
		ImportantTags:   []string{},
		MetaField:       "meta",
		TagsField:       "tags",
	}

	return l
}

func (l *ZeroLogger) formatAndCheckTags(level string, tags ...string) ([]string, bool) {
	t := append([]string{level}, l.Tags...)
	t = append(t, tags...)
	return t, l.ShouldSkipEvent(t)
}

func (l *ZeroLogger) GetLogger() *zerolog.Logger {
	if l.parent != nil {
		return l.parent.GetLogger()
	}
	return &l.Logger
}

func (l *ZeroLogger) GetCfg() *CfgLogger {
	if l.parent != nil {
		return l.parent.GetCfg()
	}
	return &l.cfg
}

func (l *ZeroLogger) GetTagsField() string {
	return l.GetCfg().TagsField
}

func (l *ZeroLogger) GetMetaField() string {
	return l.GetCfg().MetaField
}

func (l *ZeroLogger) Child() *ZeroLogger {
	nl := new(ZeroLogger)
	nl.parent = l
	copy(nl.Tags, l.Tags)

	return nl
}

func (l *ZeroLogger) With(tags []string) *ZeroLogger {
	nl := l.Child()
	nl.Tags = append(nl.Tags, tags...)
	return nl
}

func (l *ZeroLogger) WithLevel(levelTag string) *zerolog.Event {
	var level zerolog.Level
	switch levelTag {
	case DebugTag:
		level = zerolog.DebugLevel
	case InfoTag:
		level = zerolog.InfoLevel
	case WarnTag:
		level = zerolog.WarnLevel
	case FatalTag:
		level = zerolog.FatalLevel
	case PanicTag:
		level = zerolog.PanicLevel
	case ErrorTag:
		level = zerolog.ErrorLevel
	}

	return l.GetLogger().WithLevel(level)
}

func (l *ZeroLogger) Log(level string, msg string, tags ...string) {
	if t, skip := l.formatAndCheckTags(level, tags...); !skip {
		l.WithLevel(level).Strs(l.GetTagsField(), t).Msg(msg)
	}
}

func (l *ZeroLogger) Logm(level string, msg string, meta *zerolog.Event, tags ...string) {
	if t, skip := l.formatAndCheckTags(level, tags...); !skip {
		l.WithLevel(level).Dict(l.GetMetaField(), meta).Strs(l.GetTagsField(), t).Msg(msg)
	}
}

func (l *ZeroLogger) Debug(msg string, tags ...string) {
	l.Log(DebugTag, msg, tags...)
}

func (l *ZeroLogger) Debugm(msg string, meta *zerolog.Event, tags ...string) {
	l.Logm(DebugTag, msg, meta, tags...)
}

func (l *ZeroLogger) Info(msg string, tags ...string) {
	l.Log(InfoTag, msg, tags...)
}

func (l *ZeroLogger) Infom(msg string, meta *zerolog.Event, tags ...string) {
	l.Logm(InfoTag, msg, meta, tags...)
}

func (l *ZeroLogger) Warn(msg string, tags ...string) {
	l.Log(WarnTag, msg, tags...)
}

func (l *ZeroLogger) Warnm(msg string, meta *zerolog.Event, tags ...string) {
	l.Logm(WarnTag, msg, meta, tags...)
}

func (l *ZeroLogger) Error(err error, msg string, tags ...string) {
	if t, skip := l.formatAndCheckTags(ErrorTag, tags...); !skip {
		l.WithLevel(ErrorTag).Err(err).Stack().Strs(l.GetTagsField(), t).Msg(msg)
	}
}

func (l *ZeroLogger) Errorm(err error, msg string, meta *zerolog.Event, tags ...string) {
	if t, skip := l.formatAndCheckTags(ErrorTag, tags...); !skip {
		l.WithLevel(ErrorTag).Err(err).Stack().Dict(l.GetMetaField(), meta).Strs(l.GetTagsField(), t).Msg(msg)
	}
}

func (l *ZeroLogger) Fatal(msg string, tags ...string) {
	if t, skip := l.formatAndCheckTags(FatalTag, tags...); !skip {
		l.GetLogger().Fatal().Strs(l.GetTagsField(), t).Msg(msg)
	}
}

func (l *ZeroLogger) Fatalm(msg string, meta *zerolog.Event, tags ...string) {
	if t, skip := l.formatAndCheckTags(FatalTag, tags...); !skip {
		l.GetLogger().Fatal().Dict(l.GetMetaField(), meta).Strs(l.GetTagsField(), t).Msg(msg)
	}
}

func (l *ZeroLogger) Panic(msg string, tags ...string) {
	if t, skip := l.formatAndCheckTags(PanicTag, tags...); !skip {
		l.GetLogger().Panic().Strs(l.GetTagsField(), t).Msg(msg)
	}
}

func (l *ZeroLogger) Panicm(msg string, meta *zerolog.Event, tags ...string) {
	if t, skip := l.formatAndCheckTags(PanicTag, tags...); !skip {
		l.GetLogger().Panic().Dict(l.GetMetaField(), meta).Strs(l.GetTagsField(), t).Msg(msg)
	}
}

func (l *ZeroLogger) ShouldSkipEvent(tags []string) bool {
	if l.SampleRate < 1 {
		rNum := generator.Float32()
		if rNum <= l.SampleRate {
			return false
		}
	}

	if ContainImportantTags(l, tags) {
		return false
	}

	if ContainBlacklistedTags(l, tags) {
		return true
	}

	return false
}

func ContainImportantTags(l *ZeroLogger, tags []string) bool {
	for _, it := range l.cfg.ImportantTags {
		if contains(tags, it) {
			return true
		}
	}
	return false
}

func ContainBlacklistedTags(l *ZeroLogger, tags []string) bool {
	for _, it := range l.cfg.BlacklistedTags {
		if contains(tags, it) {
			return true
		}
	}
	return false
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ZeroStatsdUnmarshaler(p []byte) (StatsdData, error) {
	data := struct {
		Meta struct {
			Statsd StatsdData `json:"statsd"`
		} `json:"meta"`
	}{}
	err := json.Unmarshal(p, &data)

	return data.Meta.Statsd, err
}
