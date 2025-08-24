package otel

import "go.uber.org/fx"

var Module = fx.Module(
	"otel",
	fx.Provide(New),
)
