package user

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/user/http/handler"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"user",
	fx.Provide(
		handler.NewAuthHandler,
		handler.NewUserHandler,
	),
)
