package model

import (
	"time"
)

type OneTimeTokenModel struct {
	ID        uint64 `gorm:"primarykey"`
	UserID    uint64
	TokenHash []byte `gorm:"type:bytea"`
	TokenType string `gorm:"type:varchar(50)"`
	ExpiresAt time.Time
	CreatedAt time.Time
	UsedAt    *time.Time
}

func (*OneTimeTokenModel) TableName() string {
	return "one_time_tokens"
}
