package middleware

import "go.uber.org/fx"

var Module = fx.Module(
	"shared/http/middleware",
	fx.Provide(NewAuthMiddleware),
)
