package router

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/http/fiber/middleware"
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/http/fiber/handler"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/router"
)

func SetupContactRoutes(
	router *router.FiberRouter,
	handler *handler.ContactHandler,
	authMiddleware *middleware.AuthMiddleware,
) {
	r := router.Router()

	r.Get("/api/v1/contacts", authMiddleware.Middleware(), handler.ListContacts)
	r.Post("/api/v1/contacts", authMiddleware.Middleware(), handler.CreateContact)
	r.Put("/api/v1/contacts/:id", authMiddleware.Middleware(), handler.UpdateContact)
	r.Delete("/api/v1/contacts/:id", authMiddleware.Middleware(), handler.DeleteContact)
}
