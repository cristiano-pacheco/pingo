package http

import (
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/middleware"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/router"
	httpserver "github.com/cristiano-pacheco/pingo/pkg/httpserver/fiber"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"shared/http",
	fx.Provide(
		router.NewFiberRouter,
		httpserver.NewFiberHTTPServer,
		middleware.NewFiberErrorMiddleware,
	),
)
