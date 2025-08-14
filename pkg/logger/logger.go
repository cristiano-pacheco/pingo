package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// Logger interface mimics zerolog.Logger methods to avoid direct imports
type Logger interface {
	Debug() *zerolog.Event
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
	Fatal() *zerolog.Event
	Panic() *zerolog.Event
}

type zerologAdapter struct {
	logger   zerolog.Logger
	logLevel zerolog.Level
}

// NewLogger returns a new Logger instance with the given log level
func NewLogger(config Config) Logger {
	var level zerolog.Level
	switch config.LogLevel {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warning":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	case "fatal":
		level = zerolog.FatalLevel
	case "panic":
		level = zerolog.PanicLevel
	case "nolevel":
		level = zerolog.TraceLevel
	case "disabled":
		level = zerolog.TraceLevel
	default:
		level = zerolog.InfoLevel
	}

	zl := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return &zerologAdapter{logger: zl, logLevel: level}
}

func (z *zerologAdapter) Debug() *zerolog.Event {
	return z.logger.Debug()
}
func (z *zerologAdapter) Info() *zerolog.Event {
	return z.logger.Info()
}
func (z *zerologAdapter) Warn() *zerolog.Event {
	return z.logger.Warn()
}
func (z *zerologAdapter) Error() *zerolog.Event {
	return z.logger.Error()
}
func (z *zerologAdapter) Fatal() *zerolog.Event {
	return z.logger.Fatal()
}
func (z *zerologAdapter) Panic() *zerolog.Event {
	return z.logger.Panic()
}
