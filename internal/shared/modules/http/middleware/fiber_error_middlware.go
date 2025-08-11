package middleware

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/cristiano-pacheco/pingo/internal/shared/modules/validator"
	"github.com/cristiano-pacheco/pingo/pkg/errs"
	ut "github.com/go-playground/universal-translator"
	lib_validator "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type FiberErrorMiddleware struct {
	validate   validator.Validate
	translator ut.Translator
}

func NewFiberErrorMiddleware(
	validate validator.Validate,
	translator ut.Translator,
) *FiberErrorMiddleware {
	return &FiberErrorMiddleware{
		validate:   validate,
		translator: translator,
	}
}

func (m *FiberErrorMiddleware) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err == nil {
			return nil
		}

		// validation error flow
		if validationErrors, ok := err.(lib_validator.ValidationErrors); ok {
			var details []errs.Detail
			for _, e := range validationErrors {
				details = append(details, errs.Detail{
					Field:   m.camelToSnake(e.Field()),
					Message: e.Translate(m.translator),
				})
			}

			validationError := errs.New(
				"INVALID_ARGUMENT",
				"request has invalid fields",
				http.StatusUnprocessableEntity,
				details,
			)

			return c.Status(validationError.Status).JSON(validationError)
		}

		var customErr *errs.Error
		if errors.As(err, &customErr) {
			return c.Status(customErr.Status).JSON(customErr)
		}

		unknownError := errs.New(
			"UNKNOWN_ERROR",
			"unknown error",
			http.StatusInternalServerError,
			nil,
		)

		return c.Status(unknownError.Status).JSON(unknownError)
	}
}

func (m *FiberErrorMiddleware) camelToSnake(s string) string {
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(snake)
}
