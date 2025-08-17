package model

import (
	"time"
)

type VerificationCodeModel struct {
	ID        uint `gorm:"primarykey"`
	UserID    uint
	Code      string
	ExpiresAt time.Time
	CreatedAt time.Time
	UsedAt    *time.Time
}

func (*VerificationCodeModel) TableName() string {
	return "verification_codes"
}
