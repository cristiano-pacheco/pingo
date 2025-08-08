package translator

import (
	"errors"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"

	"github.com/cristiano-pacheco/pingo/internal/shared/modules/validator"

	lib_validator "github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

func New(v validator.Validate) ut.Translator {
	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")
	val, ok := v.(*lib_validator.Validate)
	if !ok {
		panic(errors.New("invalid validator in the translator instantiation"))
	}
	err := en_translations.RegisterDefaultTranslations(val, trans)
	if err != nil {
		panic(err)
	}
	return trans
}
