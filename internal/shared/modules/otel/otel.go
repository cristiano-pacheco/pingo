package otel

import (
	"context"
	"time"

	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/pkg/otel/trace"
	"go.uber.org/fx"
)

func New(lc fx.Lifecycle, config config.Config) trace.Trace {
	batchTimeout := time.Duration(config.OpenTelemetry.BatchTimeoutSeconds) * time.Second
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
	}

	tracer, shutdown := trace.MustNew(tc)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return shutdown(ctx)
		},
	})

	return tracer
}
