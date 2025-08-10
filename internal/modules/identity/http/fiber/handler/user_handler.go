package handler

import (
	"net/http"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/http/fiber/dto"
	"github.com/cristiano-pacheco/pingo/internal/shared/sdk/http/response"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
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
	createUserReponse := dto.CreateUserResponse{
		UserID:    1,
		FirstName: "John",
		LastName:  "Doe",
	}
	response := response.NewEnvelope(createUserReponse)
	return c.Status(http.StatusCreated).JSON(response)
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
