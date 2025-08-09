package httpserver

import (
	"context"
	"fmt"

	// ...existing code...
	"time"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	httpSwagger "github.com/swaggo/fiber-swagger"

	_ "github.com/cristiano-pacheco/pingo/docs" // imports swagger docs for API documentation
)

const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
)

type FiberHTTPServer struct {
	app *fiber.App
}

func NewFiberHTTPServer(
	corsConfig cors.Config,
	otelHandlerName string,
	isOtelEnabled bool,
	httpPort uint,
) *FiberHTTPServer {
	app := fiber.New()
	app.Use(recover.New())
	app.Use(cors.New(corsConfig))
	if isOtelEnabled {
		app.Use(otelfiber.Middleware(otelHandlerName))
	}
	// Health check
	app.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	// Metrics endpoint (not implemented: prometheus middleware not available for Fiber v2)
	// TODO: Add metrics endpoint if needed
	// Swagger
	app.Get("/swagger/*", httpSwagger.Handler())
	return &FiberHTTPServer{app: app}
}

func (s *FiberHTTPServer) App() *fiber.App {
	return s.app
}

func (s *FiberHTTPServer) Run() {
	go func() {
		if err := s.app.Listen(fmt.Sprintf(":%d", 8080)); err != nil {
			panic(err)
		}
	}()
}

func (s *FiberHTTPServer) Shutdown(ctx context.Context) error {
	return s.app.Shutdown()
}
