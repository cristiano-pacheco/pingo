package errs

import (
	"errors"
	"net/http"
)

var (
	ErrUserIsNotActivated              = errors.New("the user is not activated")
	ErrInvalidCredentials              = errors.New("invalid credentials")
	ErrInvalidToken                    = errors.New("invalid token")
	ErrInvalidAccountConfirmationToken = errors.New("invalid account confirmation token")

	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")

	ErrKeyMustBePEMEncoded = errors.New("invalid key: Key must be a PEM encoded PKCS1 or PKCS8 key")
	ErrNotRSAPrivateKey    = errors.New("key is not a valid RSA private key")

	ErrBadRequest = errors.New("bad request")

	ErrInternalServer = errors.New("internal server error")
)

func NewBadRequestError(message string) error {
	return &Error{
		Status:        http.StatusBadRequest,
		OriginalError: ErrBadRequest,
		Err: er{
			Code:    codeBadRequest,
			Message: message,
		},
	}
}
