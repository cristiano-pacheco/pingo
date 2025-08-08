package logger

import "log/slog"

type Config struct {
	IsEnabled bool
	LogLevel  LogLevel
}

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

func (l LogLevel) String() string {
	return string(l)
}

func ParseLogLevel(level LogLevel) slog.Level {
	switch level.String() {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
