package model

import "time"

type UserModel struct {
	ID                uint64 `gorm:"primarykey"`
	FirstName         string
	LastName          string
	Email             string `gorm:"uniqueIndex"`
	Status            string
	ConfirmationToken []byte `gorm:"type:bytea"`
	ConfirmedAt       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (*UserModel) TableName() string {
	return "users"
}
