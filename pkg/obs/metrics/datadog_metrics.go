package metrics

import (
	"fmt"
	"sync"

	"github.com/dhontecillas/hfw/pkg/obs/logs"
	"github.com/pkg/errors"

	"github.com/DataDog/datadog-go/statsd"
)

// Custom metric types
const (
	DataDogMetricTypeRate = iota + 100

	DataDogDefaultRate float64 = 1
)

type datadogMetrics struct {
	// read only catalog of the matrics
	catalog Catalog

	// a single client shared across the application, as is a thread
	// safe implementation
	client *statsd.Client
	rate   float64
}

// DataDogMeter is an implementation of metrics for DataDog
type DataDogMeter struct {
	log     logs.Logger
	metrics *datadogMetrics
	data    map[string]string
	dataMux sync.RWMutex
}

var _ Meter = (*DataDogMeter)(nil)

// DataDogMeterConfig has the required configuration for a DataDotMeter
type DataDogMeterConfig struct {
	StatsDAddr  string
	MetricDefs  Defs
	DefaultRate float64
}

// NewDataDogMeterBuilder creates a function to create new DataDogMeter's
func NewDataDogMeterBuilder(
	l logs.Logger, conf *DataDogMeterConfig) (MeterBuilderFn, error) {

	catalog := conf.MetricDefs.Clean(l)

	c, err := statsd.New(conf.StatsDAddr)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create datadog meter")
	}

	rate := conf.DefaultRate
	if rate <= 0.0 || rate > 1.0 {
		rate = DataDogDefaultRate
	}
	ddm := datadogMetrics{
		catalog: catalog,
		client:  c,
		rate:    rate,
	}

	return func(l logs.Logger) Meter {
		return newDataDogMeter(l, &ddm)
	}, nil
}

func newDataDogMeter(l logs.Logger, metrics *datadogMetrics) *DataDogMeter {
	return &DataDogMeter{
		log:     l,
		metrics: metrics,
		data:    make(map[string]string, len(metrics.catalog.defs)),
	}
}

// Inc increases an integer value
func (m *DataDogMeter) Inc(key string) {
	m.IncWL(key, nil)
}

// IncWL increases and integer value adding labels to this records
func (m *DataDogMeter) IncWL(key string, labels map[string]string) {
	d, _ := m.metrics.catalog.Def(key)
	if d == nil {
		m.log.WarnMsg("metric def not found").Str("name", key).Send()
		return
	}
	switch d.MetricType {
	case MetricTypeMonotonicCounter:
		err := m.metrics.client.Incr(key, m.fillLabels(d.Labels, labels), m.metrics.rate)
		if err != nil {
			m.log.ErrMsg(err, "cannot report metric").Str(
				"key", key).I64("type", int64(d.MetricType)).Send()
		}
		return
	case MetricTypeUpDownCounter:
		err := m.metrics.client.Incr(key, m.fillLabels(d.Labels, labels), m.metrics.rate)
		if err != nil {
			m.log.ErrMsg(err, "cannot report metric").Str(
				"key", key).I64("type", int64(d.MetricType)).Send()
		}
		return
	default:
		m.log.WarnMsg("metric operation (inc) invalid").Str("name", key).
			I64("type", int64(d.MetricType)).Send()
	}
}

// Dec decreases an integer value
func (m *DataDogMeter) Dec(key string) {
	m.DecWL(key, nil)
}

// DecWL decreases an integer value adding labels to this record
func (m *DataDogMeter) DecWL(key string, labels map[string]string) {
	d, _ := m.metrics.catalog.Def(key)
	if d == nil {
		m.log.WarnMsg("metric def not found").Str("name", key).Send()
		return
	}
	switch d.MetricType {
	case MetricTypeUpDownCounter:
		err := m.metrics.client.Decr(key, m.fillLabels(d.Labels, labels), m.metrics.rate)
		if err != nil {
			m.log.ErrMsg(err, "cannot report metric").Str(
				"key", key).I64("type", int64(d.MetricType)).Send()
		}
		return
	default:
		m.log.WarnMsg("invalid add operation for metric").Str("name", key).Send()
	}
}

// Add adds an int value
func (m *DataDogMeter) Add(key string, val float64) {
	m.AddWL(key, val, nil)
}

// AddWL with labels that apply only to this record
func (m *DataDogMeter) AddWL(key string, val float64, labels map[string]string) {
	d, _ := m.metrics.catalog.Def(key)
	if d == nil {
		m.log.WarnMsg("metric def not found").Str("name", key).Send()
		return
	}
	v := int64(val)
	switch d.MetricType {
	case MetricTypeMonotonicCounter:
		if v < 0 {
			m.log.WarnMsg("invalid add operation for metric").Str("name", key).Send()
			return
		}
		err := m.metrics.client.Count(key, v, m.fillLabels(d.Labels, labels), m.metrics.rate)
		if err != nil {
			m.log.ErrMsg(err, "cannot report metric").Str(
				"key", key).I64("type", int64(d.MetricType)).Send()
		}
		return
	case MetricTypeUpDownCounter:
		err := m.metrics.client.Count(key, v, m.fillLabels(d.Labels, labels), m.metrics.rate)
		if err != nil {
			m.log.ErrMsg(err, "cannot report metric").Str(
				"key", key).I64("type", int64(d.MetricType)).Send()
		}
		return
	default:
		m.log.WarnMsg("invalid add operation for metric").Str("name", key).Send()
	}
}

// Rec records a value (similar to what a Gauge would be)
func (m *DataDogMeter) Rec(key string, val float64) {
	m.RecWL(key, val, nil)
}

// RecWL with labels that apply only to this record
func (m *DataDogMeter) RecWL(key string, val float64, labels map[string]string) {
	d, _ := m.metrics.catalog.Def(key)
	if d == nil {
		m.log.WarnMsg("metric def not found").Str("name", key).Send()
		return
	}
	if d.MetricType != MetricTypeHistogram && d.MetricType != MetricTypeDistribution {
		m.log.WarnMsg("invalid add operation for metric").Str(
			"name", key).I64("type", int64(d.MetricType)).Send()
		return
	}

	if d.MetricType == MetricTypeHistogram {
		if err := m.metrics.client.Histogram(key, val,
			m.fillLabels(d.Labels, labels), m.metrics.rate); err != nil {
			m.log.Err(err, "cannot report Histogram")
		}
		return
	}
	if d.MetricType == MetricTypeDistribution {
		if err := m.metrics.client.Distribution(key, val,
			m.fillLabels(d.Labels, labels), m.metrics.rate); err != nil {
			m.log.Err(err, "cannot Distribution Histogram")
		}
		return
	}
}

// Str sets a label value for all metrics that have defined it
func (m *DataDogMeter) Str(key string, val string) {
	m.dataMux.Lock()
	m.data[key] = val
	m.dataMux.Unlock()
}

// fillLabels only fills the tags for the labels defined
func (m *DataDogMeter) fillLabels(labels []string, labelVals map[string]string) []string {
	vals := labelVals
	if vals == nil {
		vals = map[string]string{}
	}
	m.dataMux.RLock()
	defer m.dataMux.RUnlock()
	if len(m.data) == 0 {
		return nil
	}
	tags := make([]string, 0, len(labels))
	for _, l := range labels {
		v, ok := vals[l]
		if !ok {
			v = m.data[l]
		}
		tags = append(tags, fmt.Sprintf("%s:%s", l, v))
	}
	return tags
}
