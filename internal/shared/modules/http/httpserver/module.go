package httpserver

import "go.uber.org/fx"

var Module = fx.Module("sdk/httpserver",
	fx.Provide(NewHTTPServer),
)
