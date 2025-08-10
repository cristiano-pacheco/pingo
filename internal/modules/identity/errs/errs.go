package errs

import (
	"net/http"

	"github.com/cristiano-pacheco/pingo/pkg/errs"
)

var (
	ErrInvalidAccountConfirmationToken = errs.New("IDENTITY_01", "Invalid account confirmation token", http.StatusBadRequest, nil)
	ErrUserIsNotActive                 = errs.New("IDENTITY_02", "User is not active", http.StatusUnauthorized, nil)
	ErrInvalidToken                    = errs.New("IDENTITY_03", "Invalid token", http.StatusUnauthorized, nil)
	ErrInvalidCredentials              = errs.New("IDENTITY_04", "Invalid credentials", http.StatusUnauthorized, nil)
)
