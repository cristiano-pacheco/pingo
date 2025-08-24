package otel

import (
	"context"

	oteltrace "go.opentelemetry.io/otel/trace"
)

// NoopOtel implements Otel interface for testing with no-op operations.
type NoopOtel struct{}

// NewNoopOtel creates a new no-op Otel implementation for testing.
func NewNoopOtel() Otel {
	return &NoopOtel{}
}

// StartSpan implements the Otel interface with a no-op span.
func (n *NoopOtel) StartSpan(_ context.Context, name string) (context.Context, oteltrace.Span) {
	ctx := context.TODO()
	return ctx, oteltrace.SpanFromContext(ctx)
}
