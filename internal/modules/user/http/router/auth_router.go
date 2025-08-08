package router

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/user/http/handler"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/router"
)

func SetupAuthRoutes(r router.Router, h handler.AuthHandler) {
	routes := r.Router()
	routes.Post("/auth/login", h.GenerateJWTToken)
}
