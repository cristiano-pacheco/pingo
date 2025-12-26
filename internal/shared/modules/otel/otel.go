package otel

import (
	"context"
	"time"

	"github.com/cristiano-pacheco/go-otel/trace"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"go.uber.org/fx"
)

func Initialize(lc fx.Lifecycle, config config.Config) {
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
}
