package errs

import (
	"net/http"

	"github.com/cristiano-pacheco/pingo/pkg/errs"
)

var (
	ErrInvalidAccountConfirmationToken = errs.New(
		"IDENTITY_01",
		"Invalid account confirmation token",
		http.StatusBadRequest,
		nil,
	)
	ErrUserIsNotActive    = errs.New("IDENTITY_02", "User is not active", http.StatusUnauthorized, nil)
	ErrInvalidToken       = errs.New("IDENTITY_03", "Invalid token", http.StatusUnauthorized, nil)
	ErrInvalidCredentials = errs.New("IDENTITY_04", "Invalid credentials", http.StatusUnauthorized, nil)
	ErrEmailAlreadyInUse  = errs.New("IDENTITY_05", "Email already in use", http.StatusBadRequest, nil)
	ErrInvalidUserStatus  = errs.New("IDENTITY_06", "Invalid user status", http.StatusBadRequest, nil)
	ErrPasswordTooShort   = errs.New(
		"IDENTITY_07",
		"Password must be at least 8 characters long",
		http.StatusBadRequest,
		nil,
	)
	ErrPasswordNoUppercase = errs.New(
		"IDENTITY_08",
		"Password must contain at least one uppercase letter",
		http.StatusBadRequest,
		nil,
	)
	ErrPasswordNoLowercase = errs.New(
		"IDENTITY_09",
		"Password must contain at least one lowercase letter",
		http.StatusBadRequest,
		nil,
	)
	ErrPasswordNoNumber = errs.New(
		"IDENTITY_10",
		"Password must contain at least one number",
		http.StatusBadRequest,
		nil,
	)
	ErrPasswordNoSpecialChar = errs.New(
		"IDENTITY_11",
		"Password must contain at least one special character",
		http.StatusBadRequest,
		nil,
	)
	ErrUserNotFound           = errs.New("IDENTITY_12", "User not found", http.StatusNotFound, nil)
	ErrInvalidTokenType       = errs.New("IDENTITY_13", "Invalid token type", http.StatusBadRequest, nil)
	ErrUserNotInPendingStatus = errs.New("IDENTITY_14", "User is not in pending status", http.StatusBadRequest, nil)
)
