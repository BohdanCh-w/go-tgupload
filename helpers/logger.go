package helpers

import "go.uber.org/zap"

func MustLogger(opts ...zap.Option) *zap.Logger {
	logger, err := zap.NewDevelopment(opts...)
	if err != nil {
		panic(err)
	}

	return logger
}
