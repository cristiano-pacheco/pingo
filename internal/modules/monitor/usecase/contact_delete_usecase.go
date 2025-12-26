package usecase

import (
	"context"

	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/repository"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/validator"

	"github.com/cristiano-pacheco/go-otel/trace"
)

type ContactDeleteInput struct {
	ContactID uint64 `validate:"required"`
}

type ContactDeleteUseCase struct {
	contactRepository repository.ContactRepository
	validate          validator.Validate
	logger            logger.Logger
}

func NewContactDeleteUseCase(
	contactRepository repository.ContactRepository,
	validate validator.Validate,
	logger logger.Logger,
) *ContactDeleteUseCase {
	return &ContactDeleteUseCase{
		contactRepository: contactRepository,
		validate:          validate,
		logger:            logger,
	}
}

func (uc *ContactDeleteUseCase) Execute(ctx context.Context, input ContactDeleteInput) error {
	ctx, span := trace.StartSpan(ctx, "ContactDeleteUseCase.Execute")
	defer span.End()

	err := uc.validate.Struct(input)
	if err != nil {
		return err
	}

	err = uc.contactRepository.Delete(ctx, input.ContactID)
	if err != nil {
		uc.logger.Error().Msgf("error deleting contact: %v", err)
		return err
	}

	return nil
}
