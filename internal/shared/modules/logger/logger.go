package logger

import (
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/pkg/logger"
)

type Logger interface {
	logger.Logger
}

func New(config config.Config) Logger {
	logConfig := logger.Config{
		IsEnabled: config.Log.IsEnabled,
		LogLevel:  logger.LogLevel(config.Log.LogLevel),
	}
	return logger.New(logConfig)
}
