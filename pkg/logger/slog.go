package logger

import (
	"context"
	"log/slog"
	"os"
)

type slogLoggerAdapter struct {
	logger *slog.Logger
	config Config
}

func NewSlog(config Config) Logger {
	slogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return &slogLoggerAdapter{
		logger: slogger,
		config: config,
	}
}

func (l *slogLoggerAdapter) Debug(msg string, args ...any) {
	if !l.shouldLog(slog.LevelDebug) {
		return
	}
	l.logger.Debug(msg, args...)
}

func (l *slogLoggerAdapter) DebugContext(ctx context.Context, msg string, args ...any) {
	if !l.shouldLog(slog.LevelDebug) {
		return
	}
	l.logger.DebugContext(ctx, msg, args...)
}

func (l *slogLoggerAdapter) Info(msg string, args ...any) {
	if !l.shouldLog(slog.LevelInfo) {
		return
	}
	l.logger.Info(msg, args...)
}

func (l *slogLoggerAdapter) InfoContext(ctx context.Context, msg string, args ...any) {
	if !l.shouldLog(slog.LevelInfo) {
		return
	}
	l.logger.InfoContext(ctx, msg, args...)
}

func (l *slogLoggerAdapter) Warn(msg string, args ...any) {
	if !l.shouldLog(slog.LevelWarn) {
		return
	}
	l.logger.Warn(msg, args...)
}

func (l *slogLoggerAdapter) WarnContext(ctx context.Context, msg string, args ...any) {
	if !l.shouldLog(slog.LevelWarn) {
		return
	}
	l.logger.WarnContext(ctx, msg, args...)
}

func (l *slogLoggerAdapter) Error(msg string, args ...any) {
	if !l.shouldLog(slog.LevelError) {
		return
	}
	l.logger.Error(msg, args...)
}

func (l *slogLoggerAdapter) ErrorContext(ctx context.Context, msg string, args ...any) {
	if !l.shouldLog(slog.LevelError) {
		return
	}
	l.logger.ErrorContext(ctx, msg, args...)
}

func (l *slogLoggerAdapter) shouldLog(level slog.Level) bool {
	if !l.config.IsEnabled {
		return false
	}
	configLevel := ParseLogLevel(l.config.LogLevel)
	return level >= configLevel
}
