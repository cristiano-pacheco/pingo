package httpserver

import (
	"context"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	_ "github.com/cristiano-pacheco/pingo/docs" // imports swagger docs for API documentation
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/pkg/httpserver"
)

type HTTPServer struct {
	server *httpserver.HTTPServer
}

func NewHTTPServer(
	lc fx.Lifecycle,
	conf config.Config,
) *HTTPServer {
	corsConfig := httpserver.CorsConfig{
		AllowedOrigins:   conf.CORS.GetAllowedOrigins(),
		AllowedMethods:   conf.CORS.GetAllowedMethods(),
		AllowedHeaders:   conf.CORS.GetAllowedHeaders(),
		ExposedHeaders:   conf.CORS.GetExposedHeaders(),
		AllowCredentials: conf.CORS.AllowCredentials,
		MaxAge:           conf.CORS.MaxAge,
	}

	isOtelEnabled := true
	server := httpserver.NewHTTPServer(corsConfig, conf.App.Name, isOtelEnabled, conf.HTTPPort)

	httpServer := &HTTPServer{
		server: server,
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			server.Run()
			return nil
		},
		OnStop: server.Shutdown,
	})

	return httpServer
}

func (s *HTTPServer) Router() chi.Router {
	return s.server.Router()
}
