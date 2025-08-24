package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/goccy/go-json"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"

	_ "github.com/cristiano-pacheco/pingo/docs" // imports swagger docs for API documentation
)

const (
	ReadTimeout  = 10 * time.Second
	WriteTimeout = 10 * time.Second
	BodyLimit    = 1 * 1024 * 1024 // 1MB limit for REST API
)

type FiberHTTPServer struct {
	app      *fiber.App
	httpPort uint
}

func NewFiberHTTPServer(
	corsConfig cors.Config,
	appName string,
	isOtelEnabled bool,
	httpPort uint,
) *FiberHTTPServer {
	config := fiber.Config{
		EnablePrintRoutes:     true,
		DisableStartupMessage: false,
		Prefork:               false, // set to true for multi-core, false for Docker/local dev
		BodyLimit:             BodyLimit,
		ReadTimeout:           ReadTimeout,
		WriteTimeout:          WriteTimeout,
		CaseSensitive:         true,
		StrictRouting:         true,
		AppName:               appName,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
	}
	app := fiber.New(config)
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(helmet.New())
	app.Use(cors.New(corsConfig))

	prometheus := fiberprometheus.New(appName)
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	if isOtelEnabled {
		app.Use(otelfiber.Middleware())
	}

	// Health check
	app.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Swagger
	app.Get("/swagger/*", swagger.HandlerDefault) // default
	return &FiberHTTPServer{app: app, httpPort: httpPort}
}

func (s *FiberHTTPServer) App() *fiber.App {
	return s.app
}

func (s *FiberHTTPServer) Run() {
	go func() {
		if err := s.app.Listen(fmt.Sprintf(":%d", s.httpPort)); err != nil {
			panic(err)
		}
	}()
}

func (s *FiberHTTPServer) Shutdown(ctx context.Context) error {
	logger := slog.Default()
	logger.InfoContext(ctx, "Shutting down HTTP server...")
	return s.app.ShutdownWithContext(ctx)
}
