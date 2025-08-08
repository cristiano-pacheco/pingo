package errs

import "go.uber.org/fx"

var Module = fx.Module("shared/errs", fx.Provide(New))
