package enum

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
)

const (
	TokenTypeAccountConfirmation = "account_confirmation"
	TokenTypeLoginVerification   = "login_verification" // #nosec G101 -- false positive: enum token type, not credentials
	TokenTypeResetPassword       = "reset_password"
)

type TokenTypeEnum struct {
	value string
}

func NewTokenTypeEnum(value string) (TokenTypeEnum, error) {
	if value != TokenTypeLoginVerification &&
		value != TokenTypeResetPassword &&
		value != TokenTypeAccountConfirmation {
		return TokenTypeEnum{}, errs.ErrInvalidTokenType
	}
	return TokenTypeEnum{value: value}, nil
}

func (e TokenTypeEnum) String() string {
	return e.value
}
