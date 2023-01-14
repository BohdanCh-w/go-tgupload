package helpers

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func syslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("01-Jan-2006  15:04:05"))
}

func MustLogger(opts ...zap.Option) *zap.Logger {
	cfgConsole := zapcore.EncoderConfig{
		LevelKey:         "level",
		NameKey:          "logger",
		FunctionKey:      zapcore.OmitKey,
		MessageKey:       "msg",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalColorLevelEncoder,
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		TimeKey:          "time",
		EncodeTime:       syslogTimeEncoder,
		ConsoleSeparator: " | ",
	}

	core := zapcore.NewCore(zapcore.NewConsoleEncoder(cfgConsole), zapcore.Lock(os.Stdout), zap.DebugLevel)

	return zap.New(core, opts...)
}
