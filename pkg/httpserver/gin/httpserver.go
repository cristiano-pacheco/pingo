package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	_ "github.com/cristiano-pacheco/pingo/docs" // imports swagger docs for API documentation
)

const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
)

type HTTPServer struct {
	engine *gin.Engine
	server *http.Server
}

func NewHTTPServer(
	corsConfig cors.Config,
	otelHandlerName string,
	isOtelEnabled bool,
	httpPort uint,
) *HTTPServer {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	// apply CORS middleware
	engine.Use(cors.New(corsConfig))

	// Apply OpenTelemetry Gin middleware if enabled
	if isOtelEnabled {
		engine.Use(otelgin.Middleware(otelHandlerName))
	}

	// Health check
	engine.GET("/healthcheck", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Metrics endpoint
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger
	engine.GET("/swagger/*any", gin.WrapH(httpSwagger.WrapHandler))

	server := &HTTPServer{
		engine: engine,
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", httpPort),
			Handler:           engine,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
	}

	return server
}

func (s *HTTPServer) Engine() *gin.Engine {
	return s.engine
}

func (s *HTTPServer) Run() {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
}

func (s *HTTPServer) RunGraceful(ctx context.Context) error {
	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
		serverErr <- nil
	}()

	select {
	case <-ctx.Done():
		// Context cancelled, initiate graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	case err := <-serverErr:
		return err
	}
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
