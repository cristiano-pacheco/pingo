package model

import (
	"time"
)

type HTTPMonitorContactModel struct {
	HTTPMonitorID uint64 `gorm:"column:http_monitor_id"`
	ContactID     uint64 `gorm:"column:contact_id"`
	CreatedAt     time.Time
}

func (*HTTPMonitorContactModel) TableName() string {
	return "http_monitor_contacts"
}
