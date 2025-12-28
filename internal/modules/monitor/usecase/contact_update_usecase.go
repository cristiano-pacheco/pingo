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

type ContactUpdateInput struct {
	ContactID   uint64 `validate:"required"`
	Name        string `validate:"required,min=3,max=255"`
	ContactType string `validate:"required,oneof=email webhook"`
	ContactData string `validate:"required,max=500"`
	IsEnabled   bool
}

type ContactUpdateUseCase struct {
	contactValidator  monitor_validator.ContactValidator
	contactRepository repository.ContactRepository
	validate          validator.Validate
	logger            logger.Logger
}

func NewContactUpdateUseCase(
	contactValidator monitor_validator.ContactValidator,
	contactRepository repository.ContactRepository,
	validate validator.Validate,
	logger logger.Logger,
) *ContactUpdateUseCase {
	return &ContactUpdateUseCase{
		contactValidator:  contactValidator,
		contactRepository: contactRepository,
		validate:          validate,
		logger:            logger,
	}
}

func (uc *ContactUpdateUseCase) Execute(ctx context.Context, input ContactUpdateInput) error {
	ctx, span := trace.Span(ctx, "ContactUpdateUseCase.Execute")
	defer span.End()

	err := uc.validate.Struct(input)
	if err != nil {
		return err
	}

	// Validate contact type using the enum
	contactTypeEnum, err := enum.NewContactTypeEnum(input.ContactType)
	if err != nil {
		return err
	}

	// Validate contact data based on contact type
	if validationErr := uc.contactValidator.Validate(input.ContactType, input.ContactData); validationErr != nil {
		return validationErr
	}

	// Check if another contact with the same name already exists
	existingContact, err := uc.contactRepository.FindByName(ctx, input.Name)
	if err != nil && !errors.Is(err, shared_errs.ErrRecordNotFound) {
		uc.logger.Error().Msgf("error finding contact by name: %v", err)
		return err
	}

	// If a contact with the same name exists and it's not the one being updated
	if existingContact.ID != 0 && existingContact.ID != input.ContactID {
		return errs.ErrContactNameAlreadyInUse
	}

	contactModel := model.ContactModel{
		ID:          input.ContactID,
		Name:        input.Name,
		ContactType: contactTypeEnum.String(),
		ContactData: input.ContactData,
		IsEnabled:   input.IsEnabled,
	}

	_, err = uc.contactRepository.Update(ctx, contactModel)
	if err != nil {
		uc.logger.Error().Msgf("error updating contact: %v", err)
		return err
	}

	return nil
}
