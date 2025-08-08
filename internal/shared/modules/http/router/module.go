package router

import "go.uber.org/fx"

var Module = fx.Module(
	"shared/http/router",
	fx.Provide(NewRouter),
)
