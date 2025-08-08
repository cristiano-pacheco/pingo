package logger

import (
	"context"
	"log/slog"
	"os"
)

type Logger interface {
	Debug(msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
	Info(msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	Warn(msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	Error(msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

type loggerAdapter struct {
	logger *slog.Logger
	config Config
}

func New(config Config) Logger {
	slogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return &loggerAdapter{
		logger: slogger,
		config: config,
	}
}

func (l *loggerAdapter) Debug(msg string, args ...any) {
	if !l.shouldLog(slog.LevelDebug) {
		return
	}
	l.logger.Debug(msg, args...)
}

func (l *loggerAdapter) DebugContext(ctx context.Context, msg string, args ...any) {
	if !l.shouldLog(slog.LevelDebug) {
		return
	}
	l.logger.DebugContext(ctx, msg, args...)
}

func (l *loggerAdapter) Info(msg string, args ...any) {
	if !l.shouldLog(slog.LevelInfo) {
		return
	}
	l.logger.Info(msg, args...)
}

func (l *loggerAdapter) InfoContext(ctx context.Context, msg string, args ...any) {
	if !l.shouldLog(slog.LevelInfo) {
		return
	}
	l.logger.InfoContext(ctx, msg, args...)
}

func (l *loggerAdapter) Warn(msg string, args ...any) {
	if !l.shouldLog(slog.LevelWarn) {
		return
	}
	l.logger.Warn(msg, args...)
}

func (l *loggerAdapter) WarnContext(ctx context.Context, msg string, args ...any) {
	if !l.shouldLog(slog.LevelWarn) {
		return
	}
	l.logger.WarnContext(ctx, msg, args...)
}

func (l *loggerAdapter) Error(msg string, args ...any) {
	if !l.shouldLog(slog.LevelError) {
		return
	}
	l.logger.Error(msg, args...)
}

func (l *loggerAdapter) ErrorContext(ctx context.Context, msg string, args ...any) {
	if !l.shouldLog(slog.LevelError) {
		return
	}
	l.logger.ErrorContext(ctx, msg, args...)
}

func (l *loggerAdapter) shouldLog(level slog.Level) bool {
	if !l.config.IsEnabled {
		return false
	}
	configLevel := ParseLogLevel(l.config.LogLevel)
	return level >= configLevel
}
