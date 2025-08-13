package validator

import (
	"unicode"
	"unicode/utf8"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
)

type PasswordValidator interface {
	Validate(password string) error
}

type passwordValidator struct {
}

func NewPasswordValidator() PasswordValidator {
	return &passwordValidator{}
}

type passwordRequirements struct {
	hasUpper   bool
	hasLower   bool
	hasNumber  bool
	hasSpecial bool
}

func (s *passwordValidator) checkRequirements(password string) passwordRequirements {
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

func (s *passwordValidator) Validate(password string) error {
	if utf8.RuneCountInString(password) < 8 {
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
