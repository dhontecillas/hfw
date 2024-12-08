package traces

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/dhontecillas/hfw/pkg/obs/attrs"
	"github.com/dhontecillas/hfw/pkg/obs/logs"
)

type OTELTracerConfig struct {
	Host       string  `json:"host"`
	Port       int     `json:"port"`
	UseHTTP    bool    `json:"use_http"`
	SampleRate float64 `json:"sample_rate"`
}

func (c OTELTracerConfig) String() string {
	return fmt.Sprintf("%s:%d use_http: %t (sample rate: %f)",
		c.Host, c.Port, c.UseHTTP, c.SampleRate)
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
	attrList       []attribute.KeyValue
	exporter       *otlptrace.Exporter
	tracerProvider *sdktrace.TracerProvider
	tracer         trace.Tracer
	ctx            context.Context
}

// NewOTELTracerBuilder returns a function to crate new tracers.
func NewOTELTracerBuilder(ctx context.Context, l logs.Logger,
	cfg *OTELTracerConfig, serviceName string, serviceVersion string) TracerBuilderFn {
	c := cfg.Clean()
	if c.SampleRate == 0.0 {
		return NewNopTracerBuilder()
	}

	endpoint := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	var exporter *otlptrace.Exporter
	var err error
	if c.UseHTTP {
		options := make([]otlptracehttp.Option, 0, 2)
		options = append(options, otlptracehttp.WithEndpoint(endpoint))
		exporter, err = otlptracehttp.New(ctx, options...)
	} else {
		options := make([]otlptracegrpc.Option, 0, 2)
		options = append(options, otlptracegrpc.WithEndpoint(endpoint))
		// TODO: allow to have TLS config
		options = append(options, otlptracegrpc.WithInsecure())
		exporter, err = otlptracegrpc.New(ctx, options...)
	}

	if err != nil {
		l.Err(err, "cannot instantiate tracer", map[string]interface{}{
			"conf":       cfg.String(),
			"clean_conf": c.String(),
		})
	}

	if exporter == nil {
		return NewNopTracerBuilder()
	}

	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	otel.SetTextMapPropagator(prop)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(e error) {
		// TODO: we might want to "throtle" the error reporting
		// when we have repeated messagese when a OTLP backend is
		// down.
		l.Err(e, "Open Telemetry Tracer", nil)
	}))

	traceOpts := make([]sdktrace.TracerProviderOption, 0, 8)
	// we could add several span exporters, but for now a single one is enough
	traceOpts = append(traceOpts, sdktrace.WithBatcher(exporter))
	// TODO: HERE we might want to read extra trace options from config
	// to tweak how it behaves
	samplerOpt := sdktrace.WithSampler(sdktrace.AlwaysSample())
	if c.SampleRate > 0.0 && c.SampleRate < 1.0 {
		samplerOpt = sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(c.SampleRate)))
	}
	traceOpts = append(traceOpts, samplerOpt)

	res := sdkresource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(serviceVersion))

	traceOpts = append(traceOpts, sdktrace.WithResource(res))
	tracerProvider := sdktrace.NewTracerProvider(traceOpts...)
	tracer := tracerProvider.Tracer("")

	return func(log logs.Logger) Tracer {
		return &OTELTracer{
			attrList:       make([]attribute.KeyValue, 0, 16),
			exporter:       exporter,
			tracerProvider: tracerProvider,
			tracer:         tracer,
			ctx:            context.Background(),
		}
	}
}

func (t *OTELTracer) mergeAttrs(other map[string]interface{}) []attribute.KeyValue {
	merged := make([]attribute.KeyValue, 0, len(t.attrList)+len(other))
	if len(other) == 0 {
		merged = append(merged, t.attrList...)
		return merged
	}
	for _, kv := range t.attrList {
		k := string(kv.Key)
		if _, ok := other[k]; !ok {
			merged = append(merged, kv)
		}
	}
	otherAttrs := otelAttrs(other)
	merged = append(merged, otherAttrs...)
	return merged
}

// Start begins a tracer span.
func (t *OTELTracer) Start(ctx context.Context, name string, attrMap map[string]interface{}) Tracer {
	attrList := t.mergeAttrs(attrMap)
	ctx, span := t.tracer.Start(ctx, name, trace.WithAttributes(attrList...))
	return &SpanTracer{
		ctx:          ctx,
		span:         span,
		parentTracer: t,
	}
}

// End marks the end of this span.
func (t *OTELTracer) End() {
	// TODO: review this
	// can we find current main span ?
}

// TraceID returns the ID for the current trace.
func (t *OTELTracer) TraceID() string {
	return ""
}

func (t *OTELTracer) attrIdx(key string) int {
	for idx, kv := range t.attrList {
		if key == string(kv.Key) {
			return idx
		}
	}
	return -1
}

// Str adds a tag to the trace of type string.
func (t *OTELTracer) Str(key, val string) {
	if idx := t.attrIdx(key); idx >= 0 {
		t.attrList[idx].Value = attribute.StringValue(val)
		return
	}
	t.attrList = append(t.attrList, attribute.String(key, val))
	return
}

// I64 adds a tag to the trace of type int64.
func (t *OTELTracer) I64(key string, val int64) {
	if idx := t.attrIdx(key); idx >= 0 {
		t.attrList[idx].Value = attribute.Int64Value(val)
		return
	}
	t.attrList = append(t.attrList, attribute.Int64(key, val))
	return
}

// F64 adds a tag to the trace of type float64.
func (t *OTELTracer) F64(key string, val float64) {
	if idx := t.attrIdx(key); idx >= 0 {
		t.attrList[idx].Value = attribute.Float64Value(val)
		return
	}
	t.attrList = append(t.attrList, attribute.Float64(key, val))
}

// Bool adds a tag to the trace of type bool.
func (t *OTELTracer) Bool(key string, val bool) {
	if idx := t.attrIdx(key); idx >= 0 {
		t.attrList[idx].Value = attribute.BoolValue(val)
		return
	}
	t.attrList = append(t.attrList, attribute.Bool(key, val))
}

func (t *OTELTracer) SetAttrs(attrMap map[string]interface{}) {
	for k, v := range attrMap {
		attrs.SetAttr(t, k, v)
	}
}

// Err adds an error to the trace.
func (t *OTELTracer) Err(err error) {
	// no op, if it has not started
}
