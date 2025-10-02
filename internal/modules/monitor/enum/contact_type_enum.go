package enum

import "github.com/cristiano-pacheco/pingo/internal/modules/monitor/errs"

const (
	ContactTypeEmail   = "email"
	ContactTypeWebhook = "webhook"
)

type ContactTypeEnum struct {
	value string
}

func NewContactTypeEnum(value string) (ContactTypeEnum, error) {
	if value != ContactTypeEmail &&
		value != ContactTypeWebhook {
		return ContactTypeEnum{}, errs.ErrInvalidContactType
	}
	return ContactTypeEnum{value: value}, nil
}

func (e ContactTypeEnum) String() string {
	return e.value
}
