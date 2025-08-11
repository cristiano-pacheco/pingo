package httpserver

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/fx"

	_ "github.com/cristiano-pacheco/pingo/docs" // imports swagger docs for API documentation
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/middleware"
	httpserver "github.com/cristiano-pacheco/pingo/pkg/httpserver/fiber"
)

type FiberHTTPServer struct {
	server *httpserver.FiberHTTPServer
}

func NewHTTPServer(
	lc fx.Lifecycle,
	conf config.Config,
	errorMiddleware *middleware.FiberErrorMiddleware,
) *FiberHTTPServer {
	corsConfig := cors.Config{
		AllowOrigins:     conf.CORS.AllowedOrigins,
		AllowMethods:     conf.CORS.AllowedMethods,
		AllowHeaders:     conf.CORS.AllowedHeaders,
		ExposeHeaders:    conf.CORS.ExposedHeaders,
		AllowCredentials: conf.CORS.AllowCredentials,
		MaxAge:           conf.CORS.MaxAge,
	}

	isOtelEnabled := true
	server := httpserver.NewFiberHTTPServer(corsConfig, conf.App.Name, isOtelEnabled, conf.HTTPPort)

	httpServer := &FiberHTTPServer{
		server: server,
	}

	server.App().Use(errorMiddleware.Middleware())

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			server.Run()
			return nil
		},
		OnStop: server.Shutdown,
	})

	return httpServer
}

func (s *FiberHTTPServer) App() *fiber.App {
	return s.server.App()
}
