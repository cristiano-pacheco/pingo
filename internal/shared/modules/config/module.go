package config

import "go.uber.org/fx"

var Module = fx.Module("sdk/config", fx.Provide(GetConfig))
