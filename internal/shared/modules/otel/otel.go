package otel

import (
	"log/slog"

	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/pkg/otel/trace"
)

var (
	_global      Otel
	_initialized bool
)

type Otel struct {
	trace.Trace
}

func Init(cfg config.Config) {
	_global = newOtel(cfg)
	_initialized = true
}

func get() Otel {
	if !_initialized {
		//nolint:sloglint // this is a module
		slog.Error("otel not initialized")
		panic("otel not initialized")
	}
	return _global
}

func Trace() trace.Trace {
	return get().Trace
}

func newOtel(config config.Config) Otel {
	tc := trace.TracerConfig{
		AppName:      config.App.Name,
		AppVersion:   config.App.Version,
		TraceEnabled: config.Telemetry.Enabled,
		TracerVendor: config.Telemetry.TracerVendor,
		TraceURL:     config.Telemetry.TracerURL,
	}

	return Otel{
		Trace: trace.New(tc),
	}
}
