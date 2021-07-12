package go_lib_logger

import "go.uber.org/zap"

type (
	Logger struct {
		*zap.Logger
	}
)

func Create(options ...zap.Option) (*Logger, error) {
	if z, err := zap.NewProductionConfig().Build(options...); err != nil {
		return nil, err
	} else {
		return &Logger{z}, nil
	}
}
