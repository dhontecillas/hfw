package metrics

import (
	"sync"

	"github.com/dhontecillas/hfw/pkg/obs/attrs"
	"github.com/dhontecillas/hfw/pkg/obs/logs"
	"github.com/pkg/errors"

	"github.com/DataDog/datadog-go/v5/statsd"
)

// Custom metric types
const (
	DataDogMetricTypeRate = iota + 100

	DataDogDefaultRate float64 = 1
)

type datadogMetrics struct {
	// read only catalog of the matrics
	catalog MetricDefinitionList

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
	StatsDAddr  string  `json:"statsd_address"`
	DefaultRate float64 `json:"default_rate"`
}

// NewDataDogMeterBuilder creates a function to create new DataDogMeter's
func NewDataDogMeterBuilder(l logs.Logger, conf *DataDogMeterConfig,
	metricDefs MetricDefinitionList) (MeterBuilderFn, error) {

	catalog, _ := metricDefs.CleanUp()

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
		data:    make(map[string]string, len(metrics.catalog)),
	}
}

// Inc increases an integer value
func (m *DataDogMeter) Inc(key string) {
	m.IncWL(key, nil)
}

// IncWL increases and integer value adding labels to this records
func (m *DataDogMeter) IncWL(key string, labels map[string]string) {
	var err error
	var d *MetricDefinition
	d, _, err = m.metrics.catalog.Def(key,
		MetricTypeMonotonicCounter, MetricTypeUpDownCounter)
	if err == nil {
		err = m.metrics.client.Incr(key, m.fillLabels(d.Attributes, labels),
			m.metrics.rate)
	}
	if err != nil {
		m.log.Debug("cannot report metric", map[string]interface{}{
			"key":   key,
			"error": err.Error(),
		})
	}
}

// Dec decreases an integer value
func (m *DataDogMeter) Dec(key string) {
	m.DecWL(key, nil)
}

// DecWL decreases an integer value adding labels to this record
func (m *DataDogMeter) DecWL(key string, labels map[string]string) {
	var err error
	d, _, err := m.metrics.catalog.Def(key, MetricTypeUpDownCounter)
	if err == nil {
		err = m.metrics.client.Decr(key, m.fillLabels(d.Attributes, labels),
			m.metrics.rate)
	}
	if err != nil {
		m.log.Debug("cannot report metric", map[string]interface{}{
			"key":   key,
			"error": err.Error(),
		})
	}
}

// Add adds an int value
func (m *DataDogMeter) Add(key string, val int64) {
	m.AddWL(key, val, nil)
}

// AddWL with labels that apply only to this record
func (m *DataDogMeter) AddWL(key string, val int64, labels map[string]string) {
	d, _, err := m.metrics.catalog.Def(key, MetricTypeMonotonicCounter,
		MetricTypeUpDownCounter)
	if err == nil {
		err = m.metrics.client.Count(key, val,
			m.fillLabels(d.Attributes, labels), m.metrics.rate)
	}
	if err != nil {
		m.log.Debug("cannot report metric", map[string]interface{}{
			"key":   key,
			"error": err.Error(),
		})
	}
}

// Rec records a value (similar to what a Gauge would be)
func (m *DataDogMeter) Rec(key string, val float64) {
	m.RecWL(key, val, nil)
}

// RecWL with labels that apply only to this record
func (m *DataDogMeter) RecWL(key string, val float64, labels map[string]string) {
	d, _, err := m.metrics.catalog.Def(key, MetricTypeHistogram, MetricTypeDistribution)
	if err == nil {
		if d.MetricType == MetricTypeHistogram {
			err = m.metrics.client.Histogram(key, val,
				m.fillLabels(d.Attributes, labels), m.metrics.rate)
		} else {
			err = m.metrics.client.Distribution(key, val,
				m.fillLabels(d.Attributes, labels), m.metrics.rate)
		}
	}
	if err != nil {
		// how to avoid having too many of these errors in prod ?
		m.log.Debug("cannot report metric", map[string]interface{}{
			"key":   key,
			"error": err.Error(),
		})
	}
}

// Str sets a label value for all metrics that have defined it
func (m *DataDogMeter) Str(key string, val string) {
	m.dataMux.Lock()
	m.data[key] = val
	m.dataMux.Unlock()
}

// fillLabels only fills the tags for the labels defined
func (m *DataDogMeter) fillLabels(attrDefs attrs.AttrDefinitionList,
	labelVals map[string]string) []string {

	var tags []string
	extraLen := len(labelVals)

	m.dataMux.RLock()
	existingLen := len(m.data)
	tags = make([]string, 0, existingLen+extraLen)
	for _, d := range attrDefs {
		v, ok := labelVals[d.Name]
		if !ok {
			v, ok = m.data[d.Name]
		}
		if ok {
			tags = append(tags, d.Name+":"+v)
		}
	}
	m.dataMux.RUnlock()
	return tags
}
