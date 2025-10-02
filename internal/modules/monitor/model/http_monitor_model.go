package model

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type HTTPMonitorModel struct {
	ID                    uint64         `gorm:"primarykey"`
	Name                  string         `gorm:"column:name"`
	CheckTimeout          int            `gorm:"column:check_timeout"`
	FailThreshold         int16          `gorm:"column:fail_threshold"`
	CheckIntervalSeconds  int            `gorm:"column:check_interval_seconds;default:300"`
	IsEnabled             bool           `gorm:"column:is_enabled;default:true"`
	HTTPURL               string         `gorm:"column:http_url"`
	HTTPMethod            string         `gorm:"column:http_method"`
	RequestHeaders        string         `gorm:"column:request_headers;type:jsonb;default:'{}'"`
	ValidResponseStatuses pq.Int32Array  `gorm:"column:valid_response_statuses;type:integer[];default:'{200}'"`
	LastCheckedAt         sql.NullTime   `gorm:"column:last_checked_at"`
	LastStatus            sql.NullString `gorm:"column:last_status"`
	ConsecutiveFailures   int            `gorm:"column:consecutive_failures;default:0"`
	CreatedAt             time.Time      `gorm:"column:created_at"`
	UpdatedAt             time.Time      `gorm:"column:updated_at"`
}

func (*HTTPMonitorModel) TableName() string {
	return "http_monitors"
}
