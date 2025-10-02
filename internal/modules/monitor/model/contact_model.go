package model

import "time"

type ContactModel struct {
	ID          uint64 `gorm:"primarykey"`
	Name        string
	ContactType string
	ContactData string
	IsEnabled   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (*ContactModel) TableName() string {
	return "contacts"
}
