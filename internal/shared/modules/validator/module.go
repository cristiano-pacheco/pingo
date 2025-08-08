package validator

import "go.uber.org/fx"

var Module = fx.Module("kernel/validator", fx.Provide(New))
