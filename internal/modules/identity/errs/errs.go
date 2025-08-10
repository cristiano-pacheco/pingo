package errs

import (
	"net/http"

	"github.com/cristiano-pacheco/pingo/pkg/errs"
)

var (
	ErrInvalidAccountConfirmationToken = errs.New("INVALID_ACCOUNT_CONFIRMATION_TOKEN", "Invalid account confirmation token", http.StatusBadRequest, nil)
	ErrUserIsNotActive                 = errs.New("USER_IS_NOT_ACTIVE", "User is not active", http.StatusUnauthorized, nil)
	ErrInvalidToken                    = errs.New("INVALID_TOKEN", "Invalid token", http.StatusUnauthorized, nil)
	ErrInvalidCredentials              = errs.New("INVALID_CREDENTIALS", "Invalid credentials", http.StatusUnauthorized, nil)
)
