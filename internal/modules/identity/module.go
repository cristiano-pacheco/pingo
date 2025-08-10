package identity

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/http/fiber/handler"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/http/fiber/router"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"identity",
	fx.Provide(
		handler.NewAuthHandler,
		handler.NewUserHandler,
	),
	fx.Invoke(
		router.SetupUserRoutes,
		router.SetupAuthRoutes,
	),
)
