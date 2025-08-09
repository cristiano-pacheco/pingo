package errs

import (
	"net/http"

	"github.com/cristiano-pacheco/pingo/pkg/errs"
)

var (
	ErrInvalidAccountConfirmationToken = errs.New(http.StatusBadRequest, "INVALID_ACCOUNT_CONFIRMATION_TOKEN", "Invalid account confirmation token", nil)
	ErrUserIsNotActive                 = errs.New(http.StatusUnauthorized, "USER_IS_NOT_ACTIVE", "User is not active", nil)
	ErrInvalidToken                    = errs.New(http.StatusUnauthorized, "INVALID_TOKEN", "Invalid token", nil)
	ErrInvalidCredentials              = errs.New(http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid credentials", nil)
)
