package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "CreateUser"})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UpdateUser"})
}

func (h *UserHandler) ActivateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ActivateUser"})
}
