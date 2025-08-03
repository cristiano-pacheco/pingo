package model

import (
	"time"
)

type MagicTokenModel struct {
	ID        uint `gorm:"primarykey"`
	UserID    uint
	Token     []byte
	LastName  string
	ExpiresAt time.Time
	CreatedAt time.Time
	UsedAt    *time.Time
}
