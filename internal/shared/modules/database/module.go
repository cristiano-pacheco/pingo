package database

import "go.uber.org/fx"

var Module = fx.Module("kernel/database", fx.Provide(New))
