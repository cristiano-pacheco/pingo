package router

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/http/fiber/handler"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/router"
)

func SetupUserRoutes(
	router *router.FiberRouter,
	handler *handler.UserHandler,
) {
	r := router.Router()
	r.Post("/api/v1/users", handler.CreateUser)
	r.Put("/api/v1/users", handler.UpdateUser)
	r.Post("/api/v1/users/activate", handler.ActivateUser)
}
