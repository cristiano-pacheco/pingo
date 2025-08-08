package validator

import "github.com/go-playground/validator/v10"

type Validate interface {
	Struct(s interface{}) error
}

func New() Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}
