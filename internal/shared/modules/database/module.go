package database

import "go.uber.org/fx"

var Module = fx.Module("sdk/database", fx.Provide(New))
