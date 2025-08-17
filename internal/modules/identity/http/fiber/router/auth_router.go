package router

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/http/fiber/handler"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/router"
)

func SetupAuthRoutes(r *router.FiberRouter, h *handler.AuthHandler) {
	router := r.Router()
	router.Post("/api/v1/auth/login", h.Login)
	router.Post("/api/v1/auth/token", h.GenerateJWT)
}
