package model

import (
	"database/sql"
	"time"
)

type NotificationModel struct {
	ID               uint64         `gorm:"primarykey"`
	HTTPMonitorID    uint64         `gorm:"column:http_monitor_id"`
	ContactID        uint64         `gorm:"column:contact_id"`
	NotificationType string         `gorm:"column:notification_type"`
	Message          string         `gorm:"column:message"`
	Status           string         `gorm:"column:status;default:'pending'"`
	SentAt           sql.NullTime   `gorm:"column:sent_at"`
	ErrorMessage     sql.NullString `gorm:"column:error_message"`
	CreatedAt        time.Time      `gorm:"column:created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at"`
}

func (*NotificationModel) TableName() string {
	return "notifications"
}
