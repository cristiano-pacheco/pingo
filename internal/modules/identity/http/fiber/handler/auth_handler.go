package handler

import (
	"net/http"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/http/fiber/dto"
	"github.com/cristiano-pacheco/pingo/internal/shared/sdk/http/response"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// @Summary		Generate authentication token
// @Description	Authenticates user credentials and returns an access token
// @Tags		Authentication
// @Accept		json
// @Produce		json
// @Param		request	body	dto.GenerateTokenRequest	true	"Login credentials (email and password)"
// @Success		200	{object}	response.Envelope[dto.GenerateTokenResponse]	"Successfully generated token"
// @Failure		400	{object}	errs.Error	"Invalid request format or validation error"
// @Failure		401	{object}	errs.Error	"Invalid credentials"
// @Failure		404	{object}	errs.Error	"User not found"
// @Failure		500	{object}	errs.Error	"Internal server error"
// @Router		/api/v1/auth/token [post]
func (h *AuthHandler) GenerateJWTToken(c *fiber.Ctx) error {
	createUserReponse := dto.CreateUserResponse{
		UserID:    1,
		FirstName: "John",
		LastName:  "Doe",
	}
	response := response.NewEnvelope(createUserReponse)
	return c.Status(http.StatusCreated).JSON(response)
}

// @Summary		Generate authentication token
// @Description	Authenticates user credentials and returns an access token
// @Tags		Authentication
// @Accept		json
// @Produce		json
// @Param		request	body	dto.GenerateTokenRequest	true	"Login credentials (email and password)"
// @Success		200	{object}	response.Envelope[dto.GenerateTokenResponse]	"Successfully generated token"
// @Failure		400	{object}	errs.Error	"Invalid request format or validation error"
// @Failure		401	{object}	errs.Error	"Invalid credentials"
// @Failure		404	{object}	errs.Error	"User not found"
// @Failure		500	{object}	errs.Error	"Internal server error"
// @Router		/api/v1/auth/magic-link [post]
func (h *AuthHandler) SendMagicLink(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "SendMagicLink"})
}
