package validator

import (
	"unicode"
	"unicode/utf8"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
)

const (
	minimumPasswordLength = 8
)

type PasswordValidatorI interface {
	Validate(password string) error
}

type PasswordValidator struct {
}

var _ PasswordValidatorI = (*PasswordValidator)(nil)

func NewPasswordValidator() *PasswordValidator {
	return &PasswordValidator{}
}

type passwordRequirements struct {
	hasUpper   bool
	hasLower   bool
	hasNumber  bool
	hasSpecial bool
}

func (s *PasswordValidator) checkRequirements(password string) passwordRequirements {
	reqs := passwordRequirements{}

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			reqs.hasUpper = true
		case unicode.IsLower(r):
			reqs.hasLower = true
		case unicode.IsNumber(r):
			reqs.hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			reqs.hasSpecial = true
		}
	}

	return reqs
}

func (s *PasswordValidator) Validate(password string) error {
	if utf8.RuneCountInString(password) < minimumPasswordLength {
		return errs.ErrPasswordTooShort
	}

	reqs := s.checkRequirements(password)

	// Check all requirements in a single pass
	if !reqs.hasUpper {
		return errs.ErrPasswordNoUppercase
	}
	if !reqs.hasLower {
		return errs.ErrPasswordNoLowercase
	}
	if !reqs.hasNumber {
		return errs.ErrPasswordNoNumber
	}
	if !reqs.hasSpecial {
		return errs.ErrPasswordNoSpecialChar
	}

	return nil
}
