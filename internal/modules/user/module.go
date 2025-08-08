package user

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/user/http/handler"
	"github.com/cristiano-pacheco/pingo/internal/modules/user/http/router"
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
