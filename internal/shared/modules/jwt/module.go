package jwt

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"sdk/jwt",
	fx.Provide(NewParser),
)
