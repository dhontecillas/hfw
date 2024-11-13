package traces

import (
	"net/http"

	"github.com/dhontecillas/hfw/pkg/obs/logs"
)

// NopTracer is a tracer that does nothing
type NopTracer struct {
}

// NewNopTracerBuilder returns a function to crate new tracers.
func NewNopTracerBuilder() TracerBuilderFn {
	return func(log logs.Logger) Tracer {
		return &NopTracer{}
	}
}

// FromHTTPRequest checks if there is a parent trace
// to derive from in the in
func (t *NopTracer) FromHTTPRequest(r *http.Request) Tracer {
	return t
}

// Start begins a tracer span.
func (t *NopTracer) Start(name string) Tracer {
	return t
}

// End marks the end of this span.
func (t *NopTracer) End() {
}

// TraceID returns the ID for the current trace.
func (t *NopTracer) TraceID() string {
	return ""
}

// Str adds a tag to the trace of type string.
func (t *NopTracer) Str(key, val string) Tracer {
	return t
}

// I64 adds a tag to the trace of type int64.
func (t *NopTracer) I64(key string, val int64) Tracer {
	return t
}

// F64 adds a tag to the trace of type float64.
func (t *NopTracer) F64(key string, val float64) Tracer {
	return t
}

// Bool adds a tag to the trace of type bool.
func (t *NopTracer) Bool(key string, val bool) Tracer {
	return t
}

// Err adds an error to the trace.
func (t *NopTracer) Err(err error) Tracer {
	return t
}
