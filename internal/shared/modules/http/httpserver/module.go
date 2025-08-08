package httpserver

import "go.uber.org/fx"

var Module = fx.Module("kernel/httpserver",
	fx.Provide(NewHTTPServer),
)
