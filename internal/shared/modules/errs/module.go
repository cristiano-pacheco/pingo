package errs

import "go.uber.org/fx"

var Module = fx.Module("kernel/errs", fx.Provide(New))
