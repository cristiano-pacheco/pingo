package registry

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"kernel/registry",
	fx.Provide(
		fx.Annotate(
			NewPrivateKeyRegistry,
			fx.As(new(PrivateKeyRegistryI)),
		),
	),
)
