package metrics

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv/v1.27.0"

	"github.com/dhontecillas/hfw/pkg/obs/logs"
)

const (
	defaultReportingPeriod time.Duration = time.Second * 15
)

type OTELMeterConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	UseHTTP         bool   `json:"use_http"`
	ReportingPeriod string `json:"reporting_period"`
	reportingPeriod time.Duration
}

func (c OTELMeterConfig) String() string {
	return fmt.Sprintf("metrics %s:%d use_http: %t ",
		c.Host, c.Port, c.UseHTTP)
}

func (c *OTELMeterConfig) Clean() *OTELMeterConfig {
	cfg := &OTELMeterConfig{
		Host:            c.Host,
		Port:            c.Port,
		UseHTTP:         c.UseHTTP,
		ReportingPeriod: c.ReportingPeriod,
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

	d, err := time.ParseDuration(c.ReportingPeriod)
	if err != nil || d < time.Second {
		// the default repoting period is 15 secs
		d = defaultReportingPeriod
	}
	cfg.reportingPeriod = d
	return cfg
}

type OTELMetricInstrument struct {
	counter       metric.Int64Counter
	upDownCounter metric.Int64UpDownCounter
	histogram     metric.Float64Histogram
}

// OTELMetricsShared contains the runtime shared variables
// that are the exporter, the meter provider, the metric
// definitions, and the created instruments that can be accessed
// by all running instance
type OTELMetricsShared struct {
	exporter      sdkmetric.Exporter
	meterProvider *sdkmetric.MeterProvider
	meter         metric.Meter
	defs          MetricDefinitionList
	instruments   []OTELMetricInstrument
}

func (s *OTELMetricsShared) findDef(name, metricType string) int {
	for idx, md := range s.defs {
		if md.Name == name {
			if md.MetricType != metricType {
				return -1
			}
			return idx
		}
	}
	return -1
}

func (s *OTELMetricsShared) Counter(name string) metric.Int64Counter {
	idx := s.findDef(name, MetricTypeMonotonicCounter)
	if idx < 0 {
		return nil
	}
	return s.instruments[idx].counter
}

func (s *OTELMetricsShared) UpDownCounter(name string) metric.Int64UpDownCounter {
	idx := s.findDef(name, MetricTypeUpDownCounter)
	if idx < 0 {
		return nil
	}
	return s.instruments[idx].upDownCounter
}

func (s *OTELMetricsShared) Histogram(name string) metric.Float64Histogram {
	idx := s.findDef(name, MetricTypeHistogram)
	if idx < 0 {
		return nil
	}
	return s.instruments[idx].histogram
}

type OTELMeter struct {
	shared  *OTELMetricsShared
	attrMap map[string]string
	ctx     context.Context
}

// NewOTELMeterBuilder returns a function to crate new meters.
func NewOTELMeterBuilder(ctx context.Context, l logs.Logger,
	cfg *OTELMeterConfig, metricDefs MetricDefinitionList,
	serviceName string, serviceVersion string) MeterBuilderFn {

	c := cfg.Clean()

	endpoint := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	var exporter sdkmetric.Exporter
	var err error
	if c.UseHTTP {
		options := make([]otlpmetrichttp.Option, 0, 2)
		options = append(options, otlpmetrichttp.WithEndpoint(endpoint))
		exporter, err = otlpmetrichttp.New(ctx, options...)
	} else {
		options := make([]otlpmetricgrpc.Option, 0, 2)
		options = append(options, otlpmetricgrpc.WithEndpoint(endpoint))
		// TODO: allow to have TLS config
		options = append(options, otlpmetricgrpc.WithInsecure())
		exporter, err = otlpmetricgrpc.New(ctx, options...)
	}

	if err != nil {
		l.Err(err, "cannot instantiate meter", map[string]interface{}{
			"conf":       cfg.String(),
			"clean_conf": c.String(),
		})
	}

	if exporter == nil {
		b, _ := NewNopMeterBuilder()
		return b
	}

	metricOpts := make([]sdkmetric.Option, 0, 8)
	metricReader := sdkmetric.NewPeriodicReader(exporter,
		sdkmetric.WithInterval(cfg.reportingPeriod))
	metricOpts = append(metricOpts, sdkmetric.WithReader(metricReader))

	res := sdkresource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(serviceVersion))
	metricOpts = append(metricOpts, sdkmetric.WithResource(res))

	meterProvider := sdkmetric.NewMeterProvider(metricOpts...)
	// TODO: add an option to provide a name for the meter
	meter := meterProvider.Meter("default")

	mdefs, _ := metricDefs.CleanUp()
	instruments := make([]OTELMetricInstrument, 0, len(mdefs))

	for _, def := range mdefs {
		t := def.MetricType
		m := OTELMetricInstrument{}
		switch t {
		case MetricTypeMonotonicCounter:
			// TODO: check the errors
			m.counter, _ = meter.Int64Counter(def.Name)
		case MetricTypeHistogram:
			m.histogram, _ = meter.Float64Histogram(def.Name)
		case MetricTypeDistribution:
			// we do not support distribution
			continue
		case MetricTypeUpDownCounter:
			m.upDownCounter, _ = meter.Int64UpDownCounter(def.Name)
		}
		instruments = append(instruments, m)
	}

	shared := &OTELMetricsShared{
		exporter:      exporter,
		meterProvider: meterProvider,
		meter:         meter,
		defs:          mdefs,
		instruments:   instruments,
	}

	return func(log logs.Logger) Meter {
		return &OTELMeter{
			shared:  shared,
			attrMap: make(map[string]string, 16),
			ctx:     context.Background(),
		}
	}
}

func strAttr(a interface{}) string {
	switch v := a.(type) {
	case string:
		return v
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%#v", v)
	}
}

func (t *OTELMeter) mergeAttrs(other map[string]interface{}, allowedAttrs []string) []attribute.KeyValue {

	merged := make([]attribute.KeyValue, 0, len(allowedAttrs))
	for _, k := range allowedAttrs {
		if other != nil {
			if v, ok := other[k]; ok {
				merged = append(merged, attribute.String(k, strAttr(v)))
				continue
			}
		}
		if v, ok := t.attrMap[k]; ok {
			merged = append(merged, attribute.String(k, v))
		}
	}
	return merged
}

// Str adds a tag to the metric of type string.
func (t *OTELMeter) Str(key, val string) {
	t.attrMap[key] = val
}

func (t *OTELMeter) allowedAttrs(mdef *MetricDefinition) []string {
	kvs := make([]string, 0, len(mdef.Attributes))
	for _, dattr := range mdef.Attributes {
		kvs = append(kvs, dattr.Name)
	}
	return kvs
}

func (t *OTELMeter) SetAttrs(attrMap map[string]interface{}) {
	for k, v := range attrMap {
		t.attrMap[k] = strAttr(v)
	}
}

func (t *OTELMeter) Inc(key string) {
	t.IncWL(key, nil)
}

func (t *OTELMeter) Dec(key string) {
	t.DecWL(key, nil)
}

func (t *OTELMeter) Add(key string, val int64) {
	t.AddWL(key, val, nil)
}

func (t *OTELMeter) Rec(key string, val float64) {
	t.RecWL(key, val, nil)
}

func (t *OTELMeter) getAttrs(idx int, attrMap map[string]interface{}) attribute.Set {
	attrNames := t.allowedAttrs(t.shared.defs[idx])
	return attribute.NewSet(t.mergeAttrs(attrMap, attrNames)...)
}

func (t *OTELMeter) IncWL(key string, attrMap map[string]interface{}) {
	t.AddWL(key, 1, attrMap)
}

func (t *OTELMeter) DecWL(key string, attrMap map[string]interface{}) {
	t.AddWL(key, -1, attrMap)
}

func (t *OTELMeter) AddWL(key string, val int64, attrMap map[string]interface{}) {
	idx := t.shared.findDef(key, MetricTypeUpDownCounter)
	if idx == -1 {
		if val <= 0 {
			return
		}
		idx := t.shared.findDef(key, MetricTypeMonotonicCounter)
		if idx == -1 {
			return
		}
		attrList := t.getAttrs(idx, attrMap)
		t.shared.instruments[idx].upDownCounter.Add(t.ctx, val,
			metric.WithAttributeSet(attrList))
	}
	attrList := t.getAttrs(idx, attrMap)
	t.shared.instruments[idx].counter.Add(t.ctx, val,
		metric.WithAttributeSet(attrList))
}

func (t *OTELMeter) RecWL(key string, val float64, attrMap map[string]interface{}) {
	idx := t.shared.findDef(key, MetricTypeHistogram)
	if idx < 0 {
		return
	}
	attrList := t.getAttrs(idx, attrMap)
	t.shared.instruments[idx].histogram.Record(t.ctx, val,
		metric.WithAttributeSet(attrList))
}
