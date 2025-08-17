package handler

import (
	"net/http"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/http/dto"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/usecase"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authLoginUseCase *usecase.AuthLoginUseCase
}

func NewAuthHandler(
	authLoginUseCase *usecase.AuthLoginUseCase,
) *AuthHandler {
	return &AuthHandler{
		authLoginUseCase: authLoginUseCase,
	}
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
// @Router		/api/v1/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var authLoginRequest dto.AuthLoginRequest
	if err := c.BodyParser(&authLoginRequest); err != nil {
		return err
	}
	input := usecase.AuthLoginInput{
		Email:    authLoginRequest.Email,
		Password: authLoginRequest.Password,
	}
	err := h.authLoginUseCase.Execute(ctx, input)
	if err != nil {
		return err
	}
	return c.SendStatus(http.StatusNoContent)
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
