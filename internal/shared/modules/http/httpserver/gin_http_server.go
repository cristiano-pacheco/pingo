package httpserver

import (
	"context"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	_ "github.com/cristiano-pacheco/pingo/docs" // imports swagger docs for API documentation
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	httpserver "github.com/cristiano-pacheco/pingo/pkg/httpserver/gin"
)

type GinHTTPServer struct {
	server *httpserver.HTTPServer
}

func NewHTTPServer(
	lc fx.Lifecycle,
	conf config.Config,
) *GinHTTPServer {
	corsConfig := cors.Config{
		AllowOrigins:     conf.CORS.GetAllowedOrigins(),
		AllowMethods:     conf.CORS.GetAllowedMethods(),
		AllowHeaders:     conf.CORS.GetAllowedHeaders(),
		ExposeHeaders:    conf.CORS.GetExposedHeaders(),
		AllowCredentials: conf.CORS.AllowCredentials,
		MaxAge:           time.Duration(conf.CORS.MaxAge),
	}

	isOtelEnabled := true
	server := httpserver.NewHTTPServer(corsConfig, conf.App.Name, isOtelEnabled, conf.HTTPPort)

	httpServer := &GinHTTPServer{
		server: server,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			server.RunGraceful(ctx)
			return nil
		},
		OnStop: server.Shutdown,
	})

	return httpServer
}

func (s *GinHTTPServer) Engine() *gin.Engine {
	return s.server.Engine()
}
