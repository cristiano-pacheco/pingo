package jwt

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"kernel/jwt",
	fx.Provide(NewParser),
)
