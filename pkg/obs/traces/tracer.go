package traces

import (
	"net/http"

	"github.com/dhontecillas/hfw/pkg/obs/logs"
)

// Tracer defines a basic interface to create
// traces.
type Tracer interface {
	// Start begins a tracer span.
	Start(name string) Tracer
	// End marks the end of this span.
	End()
	// TraceID returns the ID for the current trace.
	TraceID() string

	// Str adds a tag to the trace of type string.
	Str(key, val string) Tracer
	// I64 adds a tag to the trace of type int64.
	I64(key string, val int64) Tracer
	// F64 adds a tag to the trace of type float64.
	F64(key string, val float64) Tracer
	// Bool adds a tag to the trace of type bool.
	Bool(key string, val bool) Tracer
	// Err adds an error to the trace.
	Err(err error) Tracer
}

// TracerBuilderFn defines the function type to create a new Tracer
type TracerBuilderFn func(log logs.Logger) Tracer

type HTTPTracer interface {
	FromHTTPRequest(r *http.Request) Tracer
}
