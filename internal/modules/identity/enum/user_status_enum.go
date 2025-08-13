package enum

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
)

const (
	UserStatusPending   = "pending"
	UserStatusActive    = "active"
	UserStatusInactive  = "inactive"
	UserStatusSuspended = "suspended"
)

type UserStatusEnum struct {
	value string
}

func (e UserStatusEnum) String() string {
	return e.value
}

func NewUserStatusEnum(value string) (UserStatusEnum, error) {
	if value != UserStatusPending &&
		value != UserStatusActive &&
		value != UserStatusInactive &&
		value != UserStatusSuspended {
		return UserStatusEnum{}, errs.ErrInvalidUserStatus
	}
	return UserStatusEnum{value: value}, nil
}
