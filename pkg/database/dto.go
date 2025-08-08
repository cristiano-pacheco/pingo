package database

import "gorm.io/gorm/logger"

type Config struct {
	Host               string
	Name               string
	User               string
	Password           string
	Port               uint
	MaxOpenConnections int
	MaxIdleConnections int
	SSLMode            bool
	PrepareSTMT        bool
	EnableLogs         bool
	LogLevel           logger.LogLevel
}
