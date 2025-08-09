package user

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/user/http/fiber/handler"
	"github.com/cristiano-pacheco/pingo/internal/modules/user/http/fiber/router"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"user",
	fx.Provide(
		handler.NewAuthHandler,
		handler.NewUserHandler,
	),
	fx.Invoke(
		router.SetupUserRoutes,
		router.SetupAuthRoutes,
	),
)
