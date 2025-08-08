package errs

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	lib_validator "github.com/go-playground/validator/v10"
)

func (em *errorMapper) mapCustomError(status int, message string) error {
	return &Error{
		Status:        status,
		OriginalError: nil,
		Err: er{
			Code:    codeUnknown,
			Message: message,
		},
	}
}

func (em *errorMapper) mapError(err error) error {
	var validationErrors lib_validator.ValidationErrors

	// validation error flow
	if errors.As(err, &validationErrors) {
		var details []detail
		for _, e := range validationErrors {
			details = append(details, detail{
				Field:   camelToSnake(e.Field()),
				Message: e.Translate(em.translator),
			})
		}

		e := Error{
			Status:        http.StatusUnprocessableEntity,
			OriginalError: err,
			Err: er{
				Code:    codeInvalidArgument,
				Message: mapMessage(codeInvalidArgument),
				Details: details,
			},
		}

		return &e
	}

	var customErr *Error
	if errors.As(err, &customErr) {
		return customErr
	}

	var (
		status = http.StatusInternalServerError
		code   string
	)

	switch {
	// Authentication
	case errors.Is(err, ErrInvalidCredentials),
		errors.Is(err, ErrUserIsNotActivated),
		errors.Is(err, ErrInvalidToken),
		errors.Is(err, ErrInvalidAccountConfirmationToken):
		status = http.StatusUnauthorized
		code = codeUnauthorized
	// Bad Request
	case errors.Is(err, ErrBadRequest):
		status = http.StatusBadRequest
		code = codeBadRequest
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
		code = codeNotFound
	default:
		// Default: internal server error
		code = codeUnknown
	}

	e := Error{
		Status:        status,
		OriginalError: err,
		Err: er{
			Code:    code,
			Message: mapMessage(code),
		},
	}

	return &e
}

func camelToSnake(s string) string {
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(snake)
}
