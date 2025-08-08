package translator

import "go.uber.org/fx"

var Module = fx.Module(
	"translator",
	fx.Provide(New),
)
