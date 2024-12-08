package traces

import (
	"context"
	"net/http"

	"github.com/dhontecillas/hfw/pkg/obs/attrs"
	"github.com/dhontecillas/hfw/pkg/obs/logs"
)

// Tracer defines a basic interface to create
// traces.
type Tracer interface {
	attrs.Attributable

	// Start begins a tracer span.
	Start(ctx context.Context, name string, attrs map[string]interface{}) Tracer
	// End marks the end of this span.
	End()

	// TraceID returns the ID for the current trace.
	TraceID() string

	// Err adds an error to the trace.
	Err(err error)
}

// TracerBuilderFn defines the function type to create a new Tracer
type TracerBuilderFn func(log logs.Logger) Tracer

type HTTPTracer interface {
	FromHTTPRequest(r *http.Request) Tracer
}
