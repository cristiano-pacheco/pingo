package handler

import (
	"net/http"
	"strconv"

	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/http/dto"
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/usecase"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/sdk/http/response"
	"github.com/gofiber/fiber/v2"
)

type ContactHandler struct {
	contactCreateUseCase *usecase.ContactCreateUseCase
	contactListUseCase   *usecase.ContactListUseCase
	contactUpdateUseCase *usecase.ContactUpdateUseCase
	contactDeleteUseCase *usecase.ContactDeleteUseCase
	logger               logger.Logger
}

func NewContactHandler(
	contactCreateUseCase *usecase.ContactCreateUseCase,
	contactListUseCase *usecase.ContactListUseCase,
	contactUpdateUseCase *usecase.ContactUpdateUseCase,
	contactDeleteUseCase *usecase.ContactDeleteUseCase,
	logger logger.Logger,
) *ContactHandler {
	return &ContactHandler{
		contactCreateUseCase: contactCreateUseCase,
		contactListUseCase:   contactListUseCase,
		contactUpdateUseCase: contactUpdateUseCase,
		contactDeleteUseCase: contactDeleteUseCase,
		logger:               logger,
	}
}

// @Summary		List contacts
// @Description	Retrieves all contacts
// @Tags		Contacts
// @Accept		json
// @Produce		json
// @Security 	BearerAuth
// @Success		200	{object}	response.Envelope[[]dto.ContactResponse]	"Successfully retrieved contacts"
// @Failure		401	{object}	errs.Error	"Invalid credentials"
// @Failure		500	{object}	errs.Error	"Internal server error"
// @Router		/api/v1/contacts [get]
func (h *ContactHandler) ListContacts(c *fiber.Ctx) error {
	ctx := c.UserContext()

	output, err := h.contactListUseCase.Execute(ctx)
	if err != nil {
		h.logger.Error().Msgf("Failed to list contacts: %v", err)
		return err
	}

	contacts := make([]dto.ContactResponse, len(output.Contacts))
	for i, contact := range output.Contacts {
		contacts[i] = dto.ContactResponse{
			ContactID:   contact.ContactID,
			Name:        contact.Name,
			ContactType: contact.ContactType,
			ContactData: contact.ContactData,
			IsEnabled:   contact.IsEnabled,
		}
	}

	res := response.NewEnvelope(contacts)
	return c.Status(http.StatusOK).JSON(res)
}

// @Summary		Create contact
// @Description	Creates a new contact
// @Tags		Contacts
// @Accept		json
// @Produce		json
// @Security 	BearerAuth
// @Param		request	body	dto.CreateContactRequest	true	"Contact data"
// @Success		201	{object}	response.Envelope[dto.CreateContactResponse]	"Successfully created contact"
// @Failure		422	{object}	errs.Error	"Invalid request format or validation error"
// @Failure		401	{object}	errs.Error	"Invalid credentials"
// @Failure		500	{object}	errs.Error	"Internal server error"
// @Router		/api/v1/contacts [post]
func (h *ContactHandler) CreateContact(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var createContactRequest dto.CreateContactRequest
	if err := c.BodyParser(&createContactRequest); err != nil {
		h.logger.Error().Msgf("Failed to parse request body: %v", err)
		return err
	}

	input := usecase.ContactCreateInput{
		Name:        createContactRequest.Name,
		ContactType: createContactRequest.ContactType,
		ContactData: createContactRequest.ContactData,
	}

	output, err := h.contactCreateUseCase.Execute(ctx, input)
	if err != nil {
		h.logger.Error().Msgf("Failed to create contact: %v", err)
		return err
	}

	createContactResponse := dto.CreateContactResponse{
		ContactID:   output.ContactID,
		Name:        output.Name,
		ContactType: output.ContactType,
		ContactData: output.ContactData,
	}

	res := response.NewEnvelope(createContactResponse)
	return c.Status(http.StatusCreated).JSON(res)
}

// @Summary		Update contact
// @Description	Updates an existing contact
// @Tags		Contacts
// @Accept		json
// @Produce		json
// @Security 	BearerAuth
// @Param		id		path	int	true	"Contact ID"
// @Param		request	body	dto.UpdateContactRequest	true	"Contact data"
// @Success		204		"Successfully updated contact"
// @Failure		422	{object}	errs.Error	"Invalid request format or validation error"
// @Failure		401	{object}	errs.Error	"Invalid credentials"
// @Failure		404	{object}	errs.Error	"Contact not found"
// @Failure		500	{object}	errs.Error	"Internal server error"
// @Router		/api/v1/contacts/{id} [put]
func (h *ContactHandler) UpdateContact(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var updateContactRequest dto.UpdateContactRequest
	if err := c.BodyParser(&updateContactRequest); err != nil {
		h.logger.Error().Msgf("Failed to parse request body: %v", err)
		return err
	}

	contactIDStr := c.Params("id")
	contactID, err := strconv.ParseUint(contactIDStr, 10, 64)
	if err != nil {
		h.logger.Error().Msgf("Invalid contact ID: %v", err)
		return fiber.NewError(http.StatusBadRequest, "Invalid contact ID")
	}

	input := usecase.ContactUpdateInput{
		ContactID:   contactID,
		Name:        updateContactRequest.Name,
		ContactType: updateContactRequest.ContactType,
		ContactData: updateContactRequest.ContactData,
		IsEnabled:   updateContactRequest.IsEnabled,
	}

	err = h.contactUpdateUseCase.Execute(ctx, input)
	if err != nil {
		h.logger.Error().Msgf("Failed to update contact: %v", err)
		return err
	}

	return c.SendStatus(http.StatusNoContent)
}

// @Summary		Delete contact
// @Description	Deletes an existing contact
// @Tags		Contacts
// @Accept		json
// @Produce		json
// @Security 	BearerAuth
// @Param		id	path	int	true	"Contact ID"
// @Success		204		"Successfully deleted contact"
// @Failure		401	{object}	errs.Error	"Invalid credentials"
// @Failure		404	{object}	errs.Error	"Contact not found"
// @Failure		500	{object}	errs.Error	"Internal server error"
// @Router		/api/v1/contacts/{id} [delete]
func (h *ContactHandler) DeleteContact(c *fiber.Ctx) error {
	ctx := c.UserContext()

	contactIDStr := c.Params("id")
	contactID, err := strconv.ParseUint(contactIDStr, 10, 64)
	if err != nil {
		h.logger.Error().Msgf("Invalid contact ID: %v", err)
		return fiber.NewError(http.StatusBadRequest, "Invalid contact ID")
	}

	input := usecase.ContactDeleteInput{
		ContactID: contactID,
	}

	err = h.contactDeleteUseCase.Execute(ctx, input)
	if err != nil {
		h.logger.Error().Msgf("Failed to delete contact: %v", err)
		return err
	}

	return c.SendStatus(http.StatusNoContent)
}
