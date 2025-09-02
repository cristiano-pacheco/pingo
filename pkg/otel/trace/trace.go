package trace

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	defaultBatchTimeout = 5 * time.Second
	defaultSampleRate   = 1.0
)

type Trace interface {
	StartSpan(ctx context.Context, name string) (context.Context, oteltrace.Span)
}

type trace struct {
	tracer         oteltrace.Tracer
	tracerProvider *sdktrace.TracerProvider
	exporter       sdktrace.SpanExporter
}

// MustNew returns a Trace and a shutdown function.
func MustNew(config TracerConfig) (Trace, func(context.Context) error) {
	trace, shutdown, err := newTrace(config)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize tracer: %v", err))
	}
	return trace, shutdown
}

func newTrace(config TracerConfig) (Trace, func(context.Context) error, error) {
	if err := config.Validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid configuration: %w", err)
	}

	config.setDefaults()

	res, err := createResource(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp, exp, err := newTracerProvider(config, res)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create tracer provider: %w", err)
	}

	setupGlobalTracing(tp)

	traceInstance := createTraceInstance(tp, exp, config.AppName)
	shutdown := createShutdownFunc(traceInstance)

	return traceInstance, shutdown, nil
}

// createResource creates and configures the OpenTelemetry resource
func createResource(config TracerConfig) (*resource.Resource, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.AppName),
			semconv.ServiceVersion(config.AppVersion),
		),
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// setupGlobalTracing configures global OpenTelemetry settings
func setupGlobalTracing(tp *sdktrace.TracerProvider) {
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(newPropagator())
}

// createTraceInstance creates a new trace instance
func createTraceInstance(tp *sdktrace.TracerProvider, exp sdktrace.SpanExporter, appName string) *trace {
	t := tp.Tracer(appName)
	return &trace{
		tracer:         t,
		tracerProvider: tp,
		exporter:       exp,
	}
}

// createShutdownFunc creates the shutdown function for the trace instance
func createShutdownFunc(traceInstance *trace) func(context.Context) error {
	logger := slog.Default()

	return func(ctx context.Context) error {
		var shutdownErr error

		if traceInstance.tracerProvider != nil {
			if err := traceInstance.tracerProvider.Shutdown(ctx); err != nil {
				logger.Error("Failed to shutdown tracer provider", "error", err)
				shutdownErr = fmt.Errorf("tracer provider shutdown failed: %w", err)
			} else {
				logger.Info("Tracer provider shutdown successfully...")
			}
		}

		if traceInstance.exporter != nil {
			if err := traceInstance.exporter.Shutdown(ctx); err != nil {
				logger.Error("Failed to shutdown exporter", "error", err)
				if shutdownErr != nil {
					return fmt.Errorf("multiple shutdown failures - tracer: %w, exporter: %w", shutdownErr, err)
				}
				return fmt.Errorf("exporter shutdown failed: %w", err)
			}
			logger.Info("Exporter shutdown successfully...")
		}

		return shutdownErr
	}
}

// newPropagator creates a composite text map propagator
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// newTracerProvider creates a new tracer provider with the given configuration
func newTracerProvider(
	config TracerConfig,
	res *resource.Resource,
) (*sdktrace.TracerProvider, sdktrace.SpanExporter, error) {
	if !config.TraceEnabled {
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.NeverSample()),
		)
		return tp, nil, nil
	}

	exp, err := newExporter(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Configure batch span processor options
	batchOptions := []sdktrace.BatchSpanProcessorOption{
		sdktrace.WithBatchTimeout(config.BatchTimeout),
		sdktrace.WithMaxExportBatchSize(config.MaxBatchSize),
	}

	// Configure sampling
	sampler := sdktrace.TraceIDRatioBased(config.SampleRate)
	if config.SampleRate >= defaultSampleRate {
		sampler = sdktrace.AlwaysSample()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exp, batchOptions...)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	return tp, exp, nil
}

// newExporter creates a new OTLP HTTP exporter
func newExporter(config TracerConfig) (sdktrace.SpanExporter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultBatchTimeout)
	defer cancel()

	options := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(config.TraceURL),
	}

	if config.Insecure {
		options = append(options, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP HTTP exporter: %w", err)
	}

	return exporter, nil
}

// StartSpan starts a new span with the given name
func (t *trace) StartSpan(ctx context.Context, name string) (context.Context, oteltrace.Span) {
	//nolint:spancheck // span is returned to caller who is responsible for ending it
	return t.tracer.Start(ctx, name)
}
