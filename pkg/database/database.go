package database

import (
	"fmt"
	"log/slog"
	"math"
	"net"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const defaultSlowQueryThreshold = 200 * time.Millisecond

type slogWriter struct {
	logger *slog.Logger
}

func (w *slogWriter) Printf(format string, args ...interface{}) {
	w.logger.Info(fmt.Sprintf(format, args...))
}

func OpenConnection(cfg Config) *gorm.DB {
	dsn := generateGormDatabaseDSN(cfg)
	gormConf := gorm.Config{}

	loggerConfig := logger.Config{
		SlowThreshold:             defaultSlowQueryThreshold, // Slow SQL threshold
		LogLevel:                  cfg.LogLevel,              // Log level
		IgnoreRecordNotFoundError: true,                      // Ignore ErrRecordNotFound error for logger
		ParameterizedQueries:      true,                      // Don't include params in the SQL log
		Colorful:                  false,                     // Disable color
	}

	newLogger := logger.New(
		&slogWriter{logger: slog.New(slog.NewTextHandler(os.Stdout, nil))}, // io writer
		loggerConfig,
	)

	if cfg.EnableLogs {
		gormConf.Logger = newLogger
	}

	pgconfig := postgres.Config{DSN: dsn}
	db, err := gorm.Open(postgres.New(pgconfig), &gormConf)
	if err != nil {
		panic(err)
	}

	return db
}

func generateGormDatabaseDSN(cfg Config) string {
	sslMode := "enabled"
	if !cfg.SSLMode {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf(
		"host=%s "+
			"user=%s "+
			"password=%s "+
			"dbname=%s "+
			"port=%d "+
			"sslmode=%s "+
			"TimeZone=UTC",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.Port,
		sslMode,
	)

	return dsn
}

func GeneratePostgresDatabaseDSN(cfg Config) string {
	if cfg.Port > math.MaxInt {
		panic(fmt.Sprintf("port value %d exceeds maximum int value", cfg.Port))
	}
	hostPort := net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port)))
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable&TimeZone=UTC",
		cfg.User,
		cfg.Password,
		hostPort,
		cfg.Name,
	)
}
