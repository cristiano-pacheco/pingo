package errs

import (
	ut "github.com/go-playground/universal-translator"

	"github.com/cristiano-pacheco/pingo/internal/shared/modules/validator"
)

type ErrorMapper interface {
	Map(err error) error
	MapCustomError(status int, message string) error
}

type errorMapper struct {
	validate   validator.Validate
	translator ut.Translator
}

func New(validate validator.Validate, translator ut.Translator) ErrorMapper {
	return &errorMapper{validate, translator}
}

func (em *errorMapper) Map(err error) error {
	return em.mapError(err)
}

func (em *errorMapper) MapCustomError(status int, message string) error {
	return em.mapCustomError(status, message)
}
