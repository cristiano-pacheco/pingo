package usecase

import (
	"context"

	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/repository"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"

	"github.com/cristiano-pacheco/go-otel/trace"
)

type ContactListOutput struct {
	Contacts []ContactListItem
}

type ContactListItem struct {
	ContactID   uint64
	Name        string
	ContactType string
	ContactData string
	IsEnabled   bool
}

type ContactListUseCase struct {
	contactRepository repository.ContactRepositoryI
	logger            logger.Logger
}

func NewContactListUseCase(
	contactRepository repository.ContactRepositoryI,
	logger logger.Logger,
) *ContactListUseCase {
	return &ContactListUseCase{
		contactRepository: contactRepository,
		logger:            logger,
	}
}

func (uc *ContactListUseCase) Execute(ctx context.Context) (ContactListOutput, error) {
	ctx, span := trace.Span(ctx, "ContactListUseCase.Execute")
	defer span.End()

	output := ContactListOutput{}

	contacts, err := uc.contactRepository.FindAll(ctx)
	if err != nil {
		uc.logger.Error().Msgf("error finding all contacts: %v", err)
		return output, err
	}

	output.Contacts = make([]ContactListItem, len(contacts))
	for i, contact := range contacts {
		output.Contacts[i] = ContactListItem{
			ContactID:   contact.ID,
			Name:        contact.Name,
			ContactType: contact.ContactType,
			ContactData: contact.ContactData,
			IsEnabled:   contact.IsEnabled,
		}
	}

	return output, nil
}
