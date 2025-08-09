package router

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/user/http/gin/handler"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/router"
)

func SetupAuthRoutes(r router.Router, h handler.AuthHandler) {
	router := r.Router()
	router.POST("/auth/login", h.GenerateJWTToken)
}
