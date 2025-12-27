package otel

import (
	"time"

	"github.com/cristiano-pacheco/go-otel/trace"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"go.uber.org/fx"
)

func Initialize(lc fx.Lifecycle, config config.Config) {
	batchTimeout := time.Duration(config.OpenTelemetry.BatchTimeoutSeconds) * time.Second

	var exporterType trace.ExporterType
	var err error

	if config.OpenTelemetry.ExporterType == trace.ExporterTypeHTTP {
		exporterType, err = trace.NewExporterType(trace.ExporterTypeHTTP)
		if err != nil {
			panic(err)
		}
	}

	if config.OpenTelemetry.ExporterType == trace.ExporterTypeGRPC {
		exporterType, err = trace.NewExporterType(trace.ExporterTypeGRPC)
		if err != nil {
			panic(err)
		}
	}

	if exporterType.IsZero() {
		panic("invalid exporter type for OpenTelemetry")
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
		OnStop: trace.Shutdown,
	})
}
