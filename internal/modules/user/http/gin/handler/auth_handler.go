package handler

import "github.com/gin-gonic/gin"

type AuthHandler struct {
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) GenerateJWTToken(c *gin.Context) {
	c.JSON(200, gin.H{"message": "GenerateJWTToken"})
}
