package go_lib_logger

import (
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

type ZLogger struct {
	zerolog.Logger
	parent *ZLogger
	Tags   []string
	cfg    CfgLogger
}

func InitLogger() *ZLogger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	l := &ZLogger{
		Logger: zerolog.New(os.Stdout),
	}
	l.cfg = CfgLogger{
		BlacklistedTags: []string{},
		ImportantTags:   []string{},
		MetaField:       "meta",
		TagsField:       "tags",
	}

	return l
}

func (l *ZLogger) formatAndCheckTags(level string, tags ...string) ([]string, bool) {
	t := append([]string{level}, l.Tags...)
	t = append(t, tags...)
	return t, l.ShouldSkipEvent(t)
}

func (l *ZLogger) GetLogger() *zerolog.Logger {
	if l.parent != nil {
		return l.parent.GetLogger()
	}
	return &l.Logger
}

func (l *ZLogger) GetCfg() *CfgLogger {
	if l.parent != nil {
		return l.parent.GetCfg()
	}
	return &l.cfg
}

func (l *ZLogger) GetTagsField() string {
	return l.GetCfg().TagsField
}

func (l *ZLogger) GetMetaField() string {
	return l.GetCfg().MetaField
}

func (l *ZLogger) Child() *ZLogger {
	nl := new(ZLogger)
	nl.parent = l
	copy(nl.Tags, l.Tags)

	return nl
}

func (l *ZLogger) With(tags []string) *ZLogger {
	nl := l.Child()
	nl.Tags = append(nl.Tags, tags...)
	return nl
}

func (l *ZLogger) WithLevel(levelTag string) *zerolog.Event {
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

func (l *ZLogger) Log(level string, msg string, tags ...string) {
	if t, skip := l.formatAndCheckTags(level, tags...); !skip {
		l.WithLevel(level).Strs(l.GetTagsField(), t).Msg(msg)
	}
}

func (l *ZLogger) Logm(level string, msg string, meta *zerolog.Event, tags ...string) {
	if t, skip := l.formatAndCheckTags(level, tags...); !skip {
		l.WithLevel(level).Dict(l.GetMetaField(), meta).Strs(l.GetTagsField(), t).Msg(msg)
	}
}

func (l *ZLogger) Debug(msg string, tags ...string) {
	l.Log(DebugTag, msg, tags...)
}

func (l *ZLogger) Debugm(msg string, meta *zerolog.Event, tags ...string) {
	l.Logm(DebugTag, msg, meta, tags...)
}

func (l *ZLogger) Info(msg string, tags ...string) {
	l.Log(InfoTag, msg, tags...)
}

func (l *ZLogger) Infom(msg string, meta *zerolog.Event, tags ...string) {
	l.Logm(InfoTag, msg, meta, tags...)
}

func (l *ZLogger) Warn(msg string, tags ...string) {
	l.Log(WarnTag, msg, tags...)
}

func (l *ZLogger) Warnm(msg string, meta *zerolog.Event, tags ...string) {
	l.Logm(WarnTag, msg, meta, tags...)
}

func (l *ZLogger) Error(err error, msg string, tags ...string) {
	if t, skip := l.formatAndCheckTags(ErrorTag, tags...); !skip {
		l.WithLevel(ErrorTag).Err(err).Stack().Strs(l.GetTagsField(), t).Msg(msg)
	}
}

func (l *ZLogger) Errorm(err error, msg string, meta *zerolog.Event, tags ...string) {
	if t, skip := l.formatAndCheckTags(ErrorTag, tags...); !skip {
		l.WithLevel(ErrorTag).Err(err).Stack().Dict(l.GetMetaField(), meta).Strs(l.GetTagsField(), t).Msg(msg)
	}
}

func (l *ZLogger) Fatal(msg string, tags ...string) {
	if t, skip := l.formatAndCheckTags(FatalTag, tags...); !skip {
		l.GetLogger().Fatal().Strs(l.GetTagsField(), t).Msg(msg)
	}
}

func (l *ZLogger) Fatalm(msg string, meta *zerolog.Event, tags ...string) {
	if t, skip := l.formatAndCheckTags(FatalTag, tags...); !skip {
		l.GetLogger().Fatal().Dict(l.GetMetaField(), meta).Strs(l.GetTagsField(), t).Msg(msg)
	}
}

func (l *ZLogger) Panic(msg string, tags ...string) {
	if t, skip := l.formatAndCheckTags(PanicTag, tags...); !skip {
		l.GetLogger().Panic().Strs(l.GetTagsField(), t).Msg(msg)
	}
}

func (l *ZLogger) Panicm(msg string, meta *zerolog.Event, tags ...string) {
	if t, skip := l.formatAndCheckTags(PanicTag, tags...); !skip {
		l.GetLogger().Panic().Dict(l.GetMetaField(), meta).Strs(l.GetTagsField(), t).Msg(msg)
	}
}

func (l *ZLogger) ShouldSkipEvent(tags []string) bool {
	if ContainImportantTags(l, tags) {
		return false
	}
	if ContainBlacklistedTags(l, tags) {
		return true
	}
	return false
}

func ContainImportantTags(l *ZLogger, tags []string) bool {
	for _, it := range l.cfg.ImportantTags {
		if contains(tags, it) {
			return true
		}
	}
	return false
}

func ContainBlacklistedTags(l *ZLogger, tags []string) bool {
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
