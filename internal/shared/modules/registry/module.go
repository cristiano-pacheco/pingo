package registry

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"kernel/registry",
	fx.Provide(
		NewPrivateKeyRegistry,
	),
)
