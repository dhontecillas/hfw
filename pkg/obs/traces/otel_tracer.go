package traces

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/dhontecillas/hfw/pkg/obs/logs"
)

type OTELTracerConfig struct {
	Host       string  `json:"host"`
	Port       int     `json:"port"`
	UseHTTP    bool    `json:"use_http"`
	SampleRate float64 `json:"sample_rate"`
}

func (c *OTELTracerConfig) Clean() *OTELTracerConfig {
	cfg := &OTELTracerConfig{
		Host:       c.Host,
		Port:       c.Port,
		UseHTTP:    c.UseHTTP,
		SampleRate: c.SampleRate,
	}

	if cfg.SampleRate < 0.0 {
		cfg.SampleRate = 0.0
	}
	if cfg.SampleRate > 1.0 {
		cfg.SampleRate = 1.0
	}

	if cfg.Port <= 0 || cfg.Port > 65535 {
		if cfg.UseHTTP {
			cfg.Port = 4318
		} else {
			cfg.Port = 4317
		}
	}

	if cfg.Host == "" {
		cfg.Host = "localhost"
	}

	return cfg
}

// OTELTracer is a tracer that does nothing
type OTELTracer struct {
	attrs []attribute.KeyValue
}

// NewOTELTracerBuilder returns a function to crate new tracers.
func NewOTELTracerBuilder(ctx context.Context, cfg *OTELTracerConfig) TracerBuilderFn {
	c := cfg.Clean()

	endpoint := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	var exporter sdktrace.SpanExporter
	if c.UseHTTP {
		options := make([]otlptracehttp.Option, 0, 2)
		options = append(options, otlptracehttp.WithEndpoint(endpoint))
		exporter, _ = otlptracehttp.New(ctx, options...)
		// TODO: log the error
	} else {
		options := make([]otlptracegrpc.Option, 0, 2)
		options = append(options, otlptracegrpc.WithEndpoint(endpoint))
		options = append(options, otlptracegrpc.WithInsecure())
		exporter, _ = otlptracegrpc.New(ctx, options...)
		// TODO: log the error
	}

	if exporter == nil {
		return NewNopTracerBuilder()
	}

	return func(log logs.Logger) Tracer {
		return &OTELTracer{
			attrs: make([]attribute.KeyValue, 0, 16),
		}
	}
}

// FromHTTPRequest checks if there is a parent trace
// to derive from in the in
func (t *OTELTracer) FromHTTPRequest(r *http.Request) Tracer {
	return t
}

// Start begins a tracer span.
func (t *OTELTracer) Start(name string) Tracer {
	return t
}

// End marks the end of this span.
func (t *OTELTracer) End() {
}

// TraceID returns the ID for the current trace.
func (t *OTELTracer) TraceID() string {
	return ""
}

// Str adds a tag to the trace of type string.
func (t *OTELTracer) Str(key, val string) Tracer {
	return t
}

// I64 adds a tag to the trace of type int64.
func (t *OTELTracer) I64(key string, val int64) Tracer {
	return t
}

// F64 adds a tag to the trace of type float64.
func (t *OTELTracer) F64(key string, val float64) Tracer {
	return t
}

// Bool adds a tag to the trace of type bool.
func (t *OTELTracer) Bool(key string, val bool) Tracer {
	return t
}

// Err adds an error to the trace.
func (t *OTELTracer) Err(err error) Tracer {
	return t
}
