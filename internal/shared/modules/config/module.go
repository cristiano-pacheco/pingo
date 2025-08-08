package config

import "go.uber.org/fx"

var Module = fx.Module("kernel/config", fx.Provide(GetConfig))
