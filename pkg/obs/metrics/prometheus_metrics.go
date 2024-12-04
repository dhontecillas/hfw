package metrics

import (
	"net/http"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/dhontecillas/hfw/pkg/obs/logs"
)

// Definitions of special metric types that can only be
// used with Prometheus.
const (
	PrometheusMetricTypeSummary = iota + 100
	PrometheusMetricTypeInvalid
)

// PrometheusConfig is the configuration for creating a Prometheus Meter.
type PrometheusConfig struct {
	ServerPort string `json:"port"`         // in the ":7323" form
	ServerPath string `json:"metrics_path"` // path to serve the metrics

	MetricsPrefix string `json:"prefix"`
}

type prometheusMetrics struct {
	Gauges     []prometheus.Gauge
	Counters   []prometheus.Counter
	Histograms []prometheus.Histogram // gauges that must be reported as histograsm
	Summaries  []prometheus.Summary   // these are an "extension" for prometheus

	LabeledCounters   []*prometheus.CounterVec
	LabeledGauges     []*prometheus.GaugeVec
	LabeledHistograms []*prometheus.HistogramVec
	LabeledSummaries  []*prometheus.SummaryVec

	catalog   MetricDefinitionList
	byTypeIdx map[string]int
}

// Serve starts the server from where the metrics can be scraped
func Serve(conf *PrometheusConfig) {
	http.Handle(conf.ServerPath, promhttp.Handler())
	go func() {
		// TODO: check if it would be better to panic
		_ = http.ListenAndServe(conf.ServerPort, nil)
	}()
}

func cleanupMetricName(in string) string {
	// prometheus does not accept '.' as part of metric name
	return strings.ReplaceAll(in, ".", "_")
}

// NewPrometheusMeterBuilder creates a new prometheus function
// to create prometheus meters.
func NewPrometheusMeterBuilder(log logs.Logger, conf *PrometheusConfig,
	metricDefs MetricDefinitionList) (MeterBuilderFn, error) {

	mdefs, _ := metricDefs.CleanUp()
	promMetrics := prometheusMetrics{
		catalog: mdefs,
	}
	promMetrics.byTypeIdx = make(map[string]int, len(promMetrics.catalog))

	for idx, def := range promMetrics.catalog {
		t := def.MetricType
		if t == MetricTypeDistribution {
			// if is not standard, nor "extension" metric skip
			log.Warn("skipping invalid metric type", map[string]interface{}{
				"key":  def.Name,
				"idx":  idx,
				"type": def.MetricType,
			})
			promMetrics.byTypeIdx[def.Name] = -1
			continue
		}

		if len(def.Attributes) == 0 {
			switch def.MetricType {
			case MetricTypeMonotonicCounter:
				c := prometheus.NewCounter(prometheus.CounterOpts{
					Name: cleanupMetricName(conf.MetricsPrefix + def.Name),
				})
				promMetrics.Counters = append(promMetrics.Counters, c)
				prometheus.MustRegister(c)
				promMetrics.byTypeIdx[def.Name] = len(promMetrics.Counters) - 1
			case MetricTypeUpDownCounter:
				g := prometheus.NewGauge(prometheus.GaugeOpts{
					Name: cleanupMetricName(conf.MetricsPrefix + def.Name),
				})
				promMetrics.Gauges = append(promMetrics.Gauges, g)
				prometheus.MustRegister(g)
				promMetrics.byTypeIdx[def.Name] = len(promMetrics.Gauges) - 1
			case MetricTypeHistogram:
				h := prometheus.NewHistogram(prometheus.HistogramOpts{
					Name: cleanupMetricName(conf.MetricsPrefix + def.Name),
					Buckets: []float64{100, 150, 200, 250, 300, 350,
						400, 450, 500, 600, 700, 1000},
				})
				promMetrics.Histograms = append(promMetrics.Histograms, h)
				prometheus.MustRegister(h)
				promMetrics.byTypeIdx[def.Name] = len(promMetrics.Histograms) - 1
			}
		} else {
			if len(def.Attributes) > 12 {
				log.Warn("high cardinality metric", map[string]interface{}{
					"key":       def.Name,
					"attrs_len": len(def.Attributes),
				})
			}

			switch def.MetricType {
			case MetricTypeMonotonicCounter:
				c := prometheus.NewCounterVec(prometheus.CounterOpts{
					Name: cleanupMetricName(conf.MetricsPrefix + def.Name),
				}, def.Attributes.Names())
				promMetrics.LabeledCounters = append(promMetrics.LabeledCounters, c)
				prometheus.MustRegister(c)
				promMetrics.byTypeIdx[def.Name] = len(promMetrics.LabeledCounters) - 1
			case MetricTypeUpDownCounter:
				g := prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Name: cleanupMetricName(conf.MetricsPrefix + def.Name),
				}, def.Attributes.Names())
				promMetrics.LabeledGauges = append(promMetrics.LabeledGauges, g)
				prometheus.MustRegister(g)
				promMetrics.byTypeIdx[def.Name] = len(promMetrics.LabeledGauges) - 1
			case MetricTypeHistogram:
				h := prometheus.NewHistogramVec(prometheus.HistogramOpts{
					Name: cleanupMetricName(conf.MetricsPrefix + def.Name),
					Buckets: []float64{20, 50, 100, 150, 200, 300,
						400, 500, 700, 1000},
				}, def.Attributes.Names())
				promMetrics.LabeledHistograms = append(promMetrics.LabeledHistograms, h)
				prometheus.MustRegister(h)
				promMetrics.byTypeIdx[def.Name] = len(promMetrics.LabeledHistograms) - 1
			}
		}
	}

	return func(log logs.Logger) Meter {
		return NewPrometheusMeter(log, &promMetrics)
	}, nil
}

// PrometheusMeter is the implementation of a Prometheus Meter
type PrometheusMeter struct {
	log     logs.Logger
	metrics *prometheusMetrics
	data    map[string]string
	dataMux sync.RWMutex
}

var _ Meter = (*PrometheusMeter)(nil)

// NewPrometheusMeter creates a new prometheus meter
func NewPrometheusMeter(log logs.Logger, promMetrics *prometheusMetrics) *PrometheusMeter {
	return &PrometheusMeter{
		log:     log,
		metrics: promMetrics,
		data:    map[string]string{},
	}
}

// Inc increases an integer value
func (pm *PrometheusMeter) Inc(key string) {
	pm.AddWL(key, 1.0, nil)
}

// IncWL increases and integer value adding labels to this records
func (pm *PrometheusMeter) IncWL(key string, labels map[string]interface{}) {
	pm.AddWL(key, 1.0, labels)
}

// Dec decreases an integer value
func (pm *PrometheusMeter) Dec(key string) {
	pm.AddWL(key, -1.0, nil)
}

// DecWL decreases an integer value adding labels to this record
func (pm *PrometheusMeter) DecWL(key string, labels map[string]interface{}) {
	pm.AddWL(key, -1.0, labels)
}

// Add adds an int value
func (pm *PrometheusMeter) Add(key string, val int64) {
	pm.AddWL(key, val, nil)
}

// AddWL with labels that apply only to this record
func (pm *PrometheusMeter) AddWL(key string, val int64, labels map[string]interface{}) {
	mdef, _, err := pm.metrics.catalog.Def(key, MetricTypeMonotonicCounter,
		MetricTypeUpDownCounter)
	if err != nil {
		pm.log.Debug("cannot find metric", map[string]interface{}{
			"key":   key,
			"error": err.Error(),
		})
		return
	}

	mtype := mdef.MetricType
	byTypeIdx := pm.metrics.byTypeIdx[mtype]
	if mtype == MetricTypeMonotonicCounter {
		if val < 0.0 {
			return
		}
		var c prometheus.Counter
		if len(mdef.Attributes) == 0 {
			c = pm.metrics.Counters[byTypeIdx]
		} else {
			lbls := pm.fillLabels(mdef.Attributes.Names(), labels)
			c, err = pm.metrics.LabeledCounters[byTypeIdx].GetMetricWith(lbls)
			if err != nil {
				pm.log.Err(err, "cannot get prometheus metric", map[string]interface{}{
					"key":   key,
					"type":  mdef.MetricType,
					"idx":   byTypeIdx,
					"error": err.Error(),
				})
				return
			}
		}
		c.Add(float64(val))
		return
	}

	var g prometheus.Gauge
	if len(mdef.Attributes) == 0 {
		g = pm.metrics.Gauges[byTypeIdx]
	} else {
		lbls := pm.fillLabels(mdef.Attributes.Names(), labels)
		g, err = pm.metrics.LabeledGauges[byTypeIdx].GetMetricWith(lbls)
		if err != nil {
			pm.log.Err(err, "cannot get prometheus metric", map[string]interface{}{
				"key":   key,
				"type":  mdef.MetricType,
				"idx":   byTypeIdx,
				"error": err.Error(),
			})
			return
		}
	}
	g.Add(float64(val))
}

// Rec records a value (similar to what a Gauge would be)
func (pm *PrometheusMeter) Rec(key string, val float64) {
	pm.RecWL(key, val, nil)
}

// RecWL with labels that apply only to this record
func (pm *PrometheusMeter) RecWL(key string, val float64, labels map[string]interface{}) {
	mdef, _, err := pm.metrics.catalog.Def(key, MetricTypeHistogram)
	if err != nil {
		pm.log.Debug("metric def not found", map[string]interface{}{
			"name":  key,
			"error": err.Error(),
		})
		return
	}
	byTypeIdx := pm.metrics.byTypeIdx[mdef.MetricType]
	var o prometheus.Observer
	if len(mdef.Attributes) == 0 {
		o = pm.metrics.Histograms[byTypeIdx]
	} else {
		lbls := pm.fillLabels(mdef.Attributes.Names(), labels)
		o, err = pm.metrics.LabeledHistograms[byTypeIdx].GetMetricWith(lbls)
		if err != nil {
			pm.log.Err(err, "cannot get prometheus metric", map[string]interface{}{
				"key":   key,
				"type":  mdef.MetricType,
				"idx":   byTypeIdx,
				"error": err.Error(),
			})
			return
		}
	}
	o.Observe(val)
}

// Str sets a label value for all metrics that have defined it
func (pm *PrometheusMeter) Str(key string, val string) {
	pm.dataMux.Lock()
	pm.data[key] = val
	pm.dataMux.Unlock()
}

func (pm *PrometheusMeter) SetAttrs(attrsMap map[string]interface{}) {
	pm.dataMux.Lock()
	for k, v := range attrsMap {
		pm.data[k] = strAttr(v)
	}
	pm.dataMux.Unlock()
}

func (pm *PrometheusMeter) fillLabels(labels []string, labelVals map[string]interface{}) prometheus.Labels {
	promLabels := make(prometheus.Labels, len(labels))
	pm.dataMux.RLock()
	for _, l := range labels {
		promLabels[l] = pm.data[l]
	}
	pm.dataMux.RUnlock()
	if labelVals == nil {
		return promLabels
	}
	for _, l := range labels {
		if v, ok := labelVals[l]; ok {
			promLabels[l] = strAttr(v)
		}
	}
	return promLabels
}
