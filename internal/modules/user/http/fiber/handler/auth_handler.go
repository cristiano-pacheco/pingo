package handler

import (
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) GenerateJWTToken(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "GenerateJWTToken"})
}
