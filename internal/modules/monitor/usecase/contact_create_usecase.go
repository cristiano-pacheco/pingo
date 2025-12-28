package usecase

import (
	"context"
	"errors"

	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/model"
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/repository"
	monitor_validator "github.com/cristiano-pacheco/pingo/internal/modules/monitor/validator"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/validator"

	"github.com/cristiano-pacheco/go-otel/trace"
)

type ContactCreateInput struct {
	Name        string `validate:"required,min=3,max=255"`
	ContactType string `validate:"required,oneof=email webhook"`
	ContactData string `validate:"required,max=500"`
}

type ContactCreateOutput struct {
	ContactID   uint64
	Name        string
	ContactType string
	ContactData string
}

type ContactCreateUseCase struct {
	contactValidator  monitor_validator.ContactValidator
	contactRepository repository.ContactRepository
	validate          validator.Validate
	logger            logger.Logger
}

func NewContactCreateUseCase(
	contactValidator monitor_validator.ContactValidator,
	contactRepository repository.ContactRepository,
	validate validator.Validate,
	logger logger.Logger,
) *ContactCreateUseCase {
	return &ContactCreateUseCase{
		contactValidator:  contactValidator,
		contactRepository: contactRepository,
		validate:          validate,
		logger:            logger,
	}
}

func (uc *ContactCreateUseCase) Execute(ctx context.Context, input ContactCreateInput) (ContactCreateOutput, error) {
	ctx, span := trace.Span(ctx, "ContactCreateUseCase.Execute")
	defer span.End()

	output := ContactCreateOutput{}

	err := uc.validate.Struct(input)
	if err != nil {
		return output, err
	}

	// Validate contact type using the enum
	contactTypeEnum, err := enum.NewContactTypeEnum(input.ContactType)
	if err != nil {
		return output, err
	}

	// Validate contact data based on contact type
	if validationErr := uc.contactValidator.Validate(input.ContactType, input.ContactData); validationErr != nil {
		return output, validationErr
	}

	// Check if contact with the same name already exists
	contact, err := uc.contactRepository.FindByName(ctx, input.Name)
	if err != nil && !errors.Is(err, shared_errs.ErrRecordNotFound) {
		uc.logger.Error().Msgf("error finding contact by name: %v", err)
		return output, err
	}

	if contact.ID != 0 {
		return output, errs.ErrContactNameAlreadyInUse
	}

	contactModel := model.ContactModel{
		Name:        input.Name,
		ContactType: contactTypeEnum.String(),
		ContactData: input.ContactData,
		IsEnabled:   true,
	}

	createdContact, err := uc.contactRepository.Create(ctx, contactModel)
	if err != nil {
		uc.logger.Error().Msgf("error creating contact: %v", err)
		return output, err
	}

	output = ContactCreateOutput{
		ContactID:   createdContact.ID,
		Name:        createdContact.Name,
		ContactType: createdContact.ContactType,
		ContactData: createdContact.ContactData,
	}

	return output, nil
}
