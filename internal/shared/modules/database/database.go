package database

import (
	"gorm.io/gorm"

	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/pkg/database"
)

type GoflixDB struct {
	*gorm.DB
}

func New(cfg config.Config) *GoflixDB {
	dbConfig := database.Config{
		Host:               cfg.DB.Host,
		User:               cfg.DB.User,
		Password:           cfg.DB.Password,
		Name:               cfg.DB.Name,
		Port:               cfg.DB.Port,
		MaxOpenConnections: cfg.DB.MaxOpenConnections,
		MaxIdleConnections: cfg.DB.MaxIdleConnections,
		SSLMode:            cfg.DB.SSLMode,
		PrepareSTMT:        cfg.DB.PrepareSTMT,
		EnableLogs:         cfg.DB.EnableLogs,
	}

	db := database.OpenConnection(dbConfig)

	return &GoflixDB{DB: db}
}

func NewFromGorm(db *gorm.DB) *GoflixDB {
	return &GoflixDB{db}
}
