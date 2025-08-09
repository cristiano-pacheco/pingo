package router

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/user/http/gin/handler"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/router"
)

func SetupUserRoutes(router router.Router, handler handler.UserHandler) {
	r := router.Router()
	r.POST("/users", handler.CreateUser)
	r.PUT("/users", handler.UpdateUser)
	r.POST("/users/activate", handler.ActivateUser)
}
