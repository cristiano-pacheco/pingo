package validator

import "go.uber.org/fx"

var Module = fx.Module("sdk/validator", fx.Provide(New))
