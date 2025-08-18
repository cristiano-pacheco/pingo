package handler

import (
	"net/http"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/http/dto"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/usecase"
	"github.com/cristiano-pacheco/pingo/internal/shared/sdk/http/response"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authLoginUseCase         *usecase.AuthLoginUseCase
	authGenerateTokenUseCase *usecase.AuthGenerateTokenUseCase
}

func NewAuthHandler(
	authLoginUseCase *usecase.AuthLoginUseCase,
	authGenerateTokenUseCase *usecase.AuthGenerateTokenUseCase,
) *AuthHandler {
	return &AuthHandler{
		authLoginUseCase:         authLoginUseCase,
		authGenerateTokenUseCase: authGenerateTokenUseCase,
	}
}

// @Summary		Authenticate the user
// @Description	Authenticates user credentials and send the verification code
// @Tags		Authentication
// @Accept		json
// @Produce		json
// @Param		request	body	dto.AuthLoginRequest	true	"Login credentials (email and password)"
// @Success		200	{object}	response.Envelope[dto.AuthLoginResponse]	"Successfully generated token"
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
	output, err := h.authLoginUseCase.Execute(ctx, input)
	if err != nil {
		return err
	}
	res := response.NewEnvelope(dto.AuthLoginResponse{UserID: output.UserID})
	return c.Status(http.StatusOK).JSON(res)
}

// @Summary		Generate authentication token
// @Description	Generate the JWT token for the user
// @Tags		Authentication
// @Accept		json
// @Produce		json
// @Param		request	body	dto.AuthGenerateJWTRequest	true	"Login credentials (email and password)"
// @Success		200	{object}	response.Envelope[dto.AuthGenerateJWTResponse]	"Successfully generated token"
// @Failure		400	{object}	errs.Error	"Invalid request format or validation error"
// @Failure		401	{object}	errs.Error	"Invalid credentials"
// @Failure		404	{object}	errs.Error	"User not found"
// @Failure		500	{object}	errs.Error	"Internal server error"
// @Router		/api/v1/auth/token [post]
func (h *AuthHandler) GenerateJWT(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var generateJWTRequest dto.AuthGenerateJWTRequest
	if err := c.BodyParser(&generateJWTRequest); err != nil {
		return err
	}
	input := usecase.GenerateTokenInput{
		UserID: generateJWTRequest.UserID,
		Code:   generateJWTRequest.Code,
	}
	output, err := h.authGenerateTokenUseCase.Execute(ctx, input)
	if err != nil {
		return err
	}

	generateJWTResponse := dto.AuthGenerateJWTResponse{
		Token: output.Token,
	}
	res := response.NewEnvelope(generateJWTResponse)
	return c.Status(http.StatusOK).JSON(res)
}
