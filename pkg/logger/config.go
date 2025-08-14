package logger

type Config struct {
	LogLevel LogLevel
}

type LogLevel string

const (
	LogLevelDebug    LogLevel = "debug"
	LogLevelInfo     LogLevel = "info"
	LogLevelWarn     LogLevel = "warning"
	LogLevelError    LogLevel = "error"
	LogLevelFatal    LogLevel = "fatal"
	LogLevelPanic    LogLevel = "panic"
	LogLevelNoLevel  LogLevel = "nolevel"
	LogLevelDisabled LogLevel = "disabled"
)

func (l LogLevel) String() string {
	return string(l)
}
