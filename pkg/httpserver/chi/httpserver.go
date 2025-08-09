package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	_ "github.com/cristiano-pacheco/pingo/docs" // imports swagger docs for API documentation
)

const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
)

type HTTPServer struct {
	router chi.Router
	server *http.Server
}

func NewHTTPServer(
	corsConfig CorsConfig,
	otelHandlerName string,
	isOtelEnabled bool,
	httpPort uint,
) *HTTPServer {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	if len(corsConfig.AllowedOrigins) > 0 {
		r.Use(corsMiddleware(corsConfig))
	}

	// Health check
	r.Get("/healthcheck", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Metrics endpoint
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})

	// Swagger
	r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
		httpSwagger.WrapHandler.ServeHTTP(w, r)
	})

	server := &HTTPServer{
		router: r,
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", httpPort),
			Handler:           r,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
	}

	// Apply OpenTelemetry if enabled
	if isOtelEnabled {
		server.server.Handler = otelhttp.NewHandler(r, otelHandlerName)
	}

	return server
}

func (s *HTTPServer) Router() chi.Router {
	return s.router
}

func (s *HTTPServer) Run() {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func corsMiddleware(corsConfig CorsConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				header := w.Header()
				header.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
				header.Set("Access-Control-Allow-Origin", corsConfig.AllowedOrigins[0])
				if len(corsConfig.AllowedHeaders) > 0 {
					header.Set("Access-Control-Allow-Headers", join(corsConfig.AllowedHeaders, ", "))
				}
				if corsConfig.AllowCredentials {
					header.Set("Access-Control-Allow-Credentials", "true")
				}
				if corsConfig.MaxAge > 0 {
					header.Set("Access-Control-Max-Age", strconv.Itoa(corsConfig.MaxAge))
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func join(s []string, sep string) string {
	if len(s) == 0 {
		return ""
	}
	result := s[0]
	for i := 1; i < len(s); i++ {
		result += sep + s[i]
	}
	return result
}
