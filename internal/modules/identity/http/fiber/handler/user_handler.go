package handler

import (
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "GenerateJWTToken"})
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "UpdateUser"})
}

func (h *UserHandler) ActivateUser(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "ActivateUser"})
}
