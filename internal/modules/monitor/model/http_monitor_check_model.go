package model

import (
	"database/sql"
	"time"
)

type HTTPMonitorCheckModel struct {
	ID             uint64         `gorm:"primarykey"`
	HTTPMonitorID  uint64         `gorm:"column:http_monitor_id"`
	CheckedAt      time.Time      `gorm:"column:checked_at"`
	ResponseTimeMs sql.NullInt32  `gorm:"column:response_time_ms"`
	StatusCode     sql.NullInt32  `gorm:"column:status_code"`
	Success        bool           `gorm:"column:success"`
	ErrorMessage   sql.NullString `gorm:"column:error_message"`
}

func (*HTTPMonitorCheckModel) TableName() string {
	return "http_monitor_checks"
}
