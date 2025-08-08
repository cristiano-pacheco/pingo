package router

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/user/http/handler"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/router"
)

func SetupUserRoutes(router router.Router, handler handler.UserHandler) {
	r := router.Router()
	r.Post("/users", handler.CreateUser)
	r.Put("/users", handler.UpdateUser)
	r.Post("/users/activate", handler.ActivateUser)
}
