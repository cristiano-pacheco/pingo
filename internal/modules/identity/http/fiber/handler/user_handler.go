package handler

import (
	"net/http"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/http/dto"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/usecase"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/sdk/http/response"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userCreateUseCase *usecase.UserCreateUseCase
	logger            logger.Logger
}

func NewUserHandler(userCreateUseCase *usecase.UserCreateUseCase, logger logger.Logger) *UserHandler {
	return &UserHandler{
		userCreateUseCase: userCreateUseCase,
		logger:            logger,
	}
}

// @Summary		Create user
// @Description	Creates a new user
// @Tags		Users
// @Accept		json
// @Produce		json
// @Param		request	body	dto.CreateUserRequest	true	"User data"
// @Success		201	{object}	response.Envelope[dto.CreateUserResponse]	"Successfully created user"
// @Failure		422	{object}	errs.Error	"Invalid request format or validation error"
// @Failure		500	{object}	errs.Error	"Internal server error"
// @Router		/api/v1/users [post]
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var createUserRequest dto.CreateUserRequest
	if err := c.BodyParser(&createUserRequest); err != nil {
		h.logger.Error().Msgf("Failed to parse request body: %v", err)
		return err
	}

	input := usecase.UserCreateInput{
		FirstName: createUserRequest.FirstName,
		LastName:  createUserRequest.LastName,
		Email:     createUserRequest.Email,
	}

	createUserResponse, err := h.userCreateUseCase.Execute(ctx, input)
	if err != nil {
		h.logger.Error().Msgf("Failed to create user: %v", err)
		return err
	}

	res := response.NewEnvelope(createUserResponse)
	return c.Status(http.StatusCreated).JSON(res)
}

// @Summary		Update user
// @Description	Updates an existing user
// @Tags		Users
// @Accept		json
// @Produce		json
// @Security 	BearerAuth
// @Param		id		path	integer		true	"User ID"
// @Param		request	body	dto.UpdateUserRequest	true	"User data"
// @Success		204		"Successfully updated user"
// @Failure		422	{object}	errs.Error	"Invalid request format or validation error"
// @Failure		401	{object}	errs.Error	"Invalid credentials"
// @Failure		404	{object}	errs.Error	"User not found"
// @Failure		500	{object}	errs.Error	"Internal server error"
// @Router		/api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "UpdateUser"})
}

// @Summary		Activate user
// @Description	Activates an existing user
// @Tags		Users
// @Accept		json
// @Produce		json
// @Param		request	body	dto.ActivateUserRequest	true	"User data"
// @Success		204		"Successfully activated user"
// @Failure		400	{object}	errs.Error	"Invalid request format or validation error"
// @Failure		500	{object}	errs.Error	"Internal server error"
// @Router		/api/v1/users/activate [post]
func (h *UserHandler) ActivateUser(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "ActivateUser"})
}
