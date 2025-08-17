package model

import (
	"time"
)

type VerificationCodeModel struct {
	ID        uint64 `gorm:"primarykey"`
	UserID    uint64
	Code      string
	ExpiresAt time.Time
	CreatedAt time.Time
	UsedAt    *time.Time
}

func (*VerificationCodeModel) TableName() string {
	return "verification_codes"
}
