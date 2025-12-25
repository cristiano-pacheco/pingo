package otel

import (
	"context"
	"time"

	"github.com/cristiano-pacheco/go-otel/trace"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

type Otel interface {
	StartSpan(ctx context.Context, name string) (context.Context, oteltrace.Span)
}

func New(lc fx.Lifecycle, config config.Config) Otel {
	batchTimeout := time.Duration(config.OpenTelemetry.BatchTimeoutSeconds) * time.Second

	exporterType, err := trace.NewExporterType(trace.ExporterTypeHTTP)
	if err != nil {
		panic(err)
	}

	tc := trace.TracerConfig{
		AppName:      config.App.Name,
		AppVersion:   config.App.Version,
		TraceEnabled: config.OpenTelemetry.Enabled,
		TracerVendor: config.OpenTelemetry.TracerVendor,
		TraceURL:     config.OpenTelemetry.TracerURL,
		BatchTimeout: batchTimeout,
		MaxBatchSize: config.OpenTelemetry.MaxBatchSize,
		Insecure:     config.OpenTelemetry.Insecure,
		SampleRate:   config.OpenTelemetry.SampleRate,
		ExporterType: exporterType,
	}

	trace.MustInitialize(tc)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return trace.Shutdown(ctx)
		},
	})

	return &otelWrapper{}
}

type otelWrapper struct{}

func (o *otelWrapper) StartSpan(ctx context.Context, name string) (context.Context, oteltrace.Span) {
	return trace.StartSpan(ctx, name)
}
