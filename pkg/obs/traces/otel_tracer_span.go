package traces

import (
	"context"

	"github.com/dhontecillas/hfw/pkg/obs/attrs"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var nilSpanTracer Tracer = (*SpanTracer)(nil)
var nilAttributableTracer attrs.Attributable = (*SpanTracer)(nil)

type SpanTracer struct {
	ctx          context.Context
	span         trace.Span
	parentTracer *OTELTracer
}

// Start begins a tracer span.
func (s *SpanTracer) Start(ctx context.Context, name string, attrMap map[string]interface{}) Tracer {
	// TODO: should I use the saved context, or the provided context ?
	attrList := s.parentTracer.mergeAttrs(attrMap)
	ctx, span := s.parentTracer.tracer.Start(ctx, name, trace.WithAttributes(attrList...))
	return &SpanTracer{
		ctx:          ctx,
		span:         span,
		parentTracer: s.parentTracer,
	}
}

// End marks the end of this span.
func (s *SpanTracer) End() {
	s.span.End()
}

// TraceID returns the ID for the current trace.
func (s *SpanTracer) TraceID() string {
	sc := s.span.SpanContext()
	spanID := sc.SpanID().String()
	traceID := sc.TraceID().String()
	if len(spanID) == 0 && len(traceID) == 0 {
		return ""
	}
	return traceID + "::" + spanID
}

// Str adds a tag to the trace of type string.
func (s *SpanTracer) Str(key, val string) {
	s.span.SetAttributes(attribute.String(key, val))
}

// I64 adds a tag to the trace of type int64.
func (s *SpanTracer) I64(key string, val int64) {
	s.span.SetAttributes(attribute.Int64(key, val))
}

// F64 adds a tag to the trace of type float64.
func (s *SpanTracer) F64(key string, val float64) {
	s.span.SetAttributes(attribute.Float64(key, val))
}

// Bool adds a tag to the trace of type bool.
func (s *SpanTracer) Bool(key string, val bool) {
	s.span.SetAttributes(attribute.Bool(key, val))
}

func (s *SpanTracer) SetAttrs(attrMap map[string]interface{}) {
	attrs.ApplyAttrs(s, attrMap)
}

// Err adds an error to the trace.
func (s *SpanTracer) Err(err error) {
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}
