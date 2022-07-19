package metrics

import (
	"fmt"
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
	ServerPort string // in the ":7323" form
	ServerPath string // path to serve the metrics

	MetricDefinitions Defs
	MetricsPrefix     string
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

	catalog   Catalog
	byTypeIdx map[int]int
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
func NewPrometheusMeterBuilder(
	log logs.Logger, conf *PrometheusConfig) (MeterBuilderFn, error) {

	promMetrics := prometheusMetrics{
		catalog: conf.MetricDefinitions.Clean(log),
	}
	promMetrics.byTypeIdx = make(map[int]int, len(promMetrics.catalog.defs))

	for idx, def := range promMetrics.catalog.defs {
		t := def.MetricType
		if t < MetricTypeMonotonicCounter || t >= MetricTypeInvalid || t == MetricTypeDistribution {
			if t < PrometheusMetricTypeSummary || t >= PrometheusMetricTypeInvalid {
				// if is not standard, nor "extension" metric skip
				log.Warn(fmt.Sprintf("skipping %s for with invalid metric type %d",
					def.Name, def.MetricType))
				promMetrics.byTypeIdx[idx] = -1
				continue
			}
		}

		if len(def.Labels) == 0 {
			switch def.MetricType {
			case MetricTypeMonotonicCounter:
				c := prometheus.NewCounter(prometheus.CounterOpts{
					Name: cleanupMetricName(conf.MetricsPrefix + def.Name),
				})
				promMetrics.Counters = append(promMetrics.Counters, c)
				prometheus.MustRegister(c)
				promMetrics.byTypeIdx[idx] = len(promMetrics.Counters) - 1
			case MetricTypeUpDownCounter:
				g := prometheus.NewGauge(prometheus.GaugeOpts{
					Name: cleanupMetricName(conf.MetricsPrefix + def.Name),
				})
				promMetrics.Gauges = append(promMetrics.Gauges, g)
				prometheus.MustRegister(g)
				promMetrics.byTypeIdx[idx] = len(promMetrics.Gauges) - 1
			case MetricTypeHistogram:
				h := prometheus.NewHistogram(prometheus.HistogramOpts{
					Name: cleanupMetricName(conf.MetricsPrefix + def.Name),
					Buckets: []float64{100, 150, 200, 250, 300, 350,
						400, 450, 500, 600, 700, 1000},
				})
				promMetrics.Histograms = append(promMetrics.Histograms, h)
				prometheus.MustRegister(h)
				promMetrics.byTypeIdx[idx] = len(promMetrics.Histograms) - 1
			}
		} else {
			if len(def.Labels) > 12 {
				log.Warn(fmt.Sprintf("high cardinality metric: %s , labels len %d",
					def.Name, len(def.Labels)))
			}
			switch def.MetricType {
			case MetricTypeMonotonicCounter:
				c := prometheus.NewCounterVec(prometheus.CounterOpts{
					Name: cleanupMetricName(conf.MetricsPrefix + def.Name),
				}, def.Labels)
				promMetrics.LabeledCounters = append(promMetrics.LabeledCounters, c)
				prometheus.MustRegister(c)
				promMetrics.byTypeIdx[idx] = len(promMetrics.LabeledCounters) - 1
			case MetricTypeUpDownCounter:
				g := prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Name: cleanupMetricName(conf.MetricsPrefix + def.Name),
				}, def.Labels)
				promMetrics.LabeledGauges = append(promMetrics.LabeledGauges, g)
				prometheus.MustRegister(g)
				promMetrics.byTypeIdx[idx] = len(promMetrics.LabeledGauges) - 1
			case MetricTypeHistogram:
				h := prometheus.NewHistogramVec(prometheus.HistogramOpts{
					Name: cleanupMetricName(conf.MetricsPrefix + def.Name),
					Buckets: []float64{20, 50, 100, 150, 200, 300,
						400, 500, 700, 1000},
				}, def.Labels)
				promMetrics.LabeledHistograms = append(promMetrics.LabeledHistograms, h)
				prometheus.MustRegister(h)
				promMetrics.byTypeIdx[idx] = len(promMetrics.LabeledHistograms) - 1
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
func (pm *PrometheusMeter) IncWL(key string, labels map[string]string) {
	pm.AddWL(key, 1.0, labels)
}

// Dec decreases an integer value
func (pm *PrometheusMeter) Dec(key string) {
	pm.AddWL(key, -1.0, nil)
}

// DecWL decreases an integer value adding labels to this record
func (pm *PrometheusMeter) DecWL(key string, labels map[string]string) {
	pm.AddWL(key, -1.0, labels)
}

// Add adds an int value
func (pm *PrometheusMeter) Add(key string, val float64) {
	pm.AddWL(key, val, nil)
}

// AddWL with labels that apply only to this record
func (pm *PrometheusMeter) AddWL(key string, val float64, labels map[string]string) {
	var err error
	mdef, idx := pm.metrics.catalog.Def(key)
	if mdef == nil {
		pm.log.WarnMsg("whop: metric def not found").Str("name", key).Send()
		return
	}
	mtype := mdef.MetricType
	if mtype != MetricTypeMonotonicCounter && mtype != MetricTypeUpDownCounter {
		pm.log.Err(fmt.Errorf("bad metric operation"),
			fmt.Sprintf("cannot use Add with %s metric", key))
		return
	}
	byTypeIdx := pm.metrics.byTypeIdx[idx]
	if mtype == MetricTypeMonotonicCounter {
		if val < 0.0 {
			pm.log.Err(fmt.Errorf("bad Metric Value"),
				fmt.Sprintf("counter %s can only add positive values", mdef.Name))
			return
		}
		var c prometheus.Counter
		if len(mdef.Labels) == 0 {
			c = pm.metrics.Counters[byTypeIdx]
		} else {
			lbls := pm.fillLabels(mdef.Labels, labels)
			c, err = pm.metrics.LabeledCounters[byTypeIdx].GetMetricWith(lbls)
			if err != nil {
				pm.log.Err(err, fmt.Sprintf("cannot get metric %s", key))
				return
			}
		}
		c.Add(val)
		return
	}
	var g prometheus.Gauge
	if len(mdef.Labels) == 0 {
		g = pm.metrics.Gauges[byTypeIdx]
	} else {
		lbls := pm.fillLabels(mdef.Labels, labels)
		g, err = pm.metrics.LabeledGauges[byTypeIdx].GetMetricWith(lbls)
		if err != nil {
			pm.log.Err(err,
				fmt.Sprintf("Cannot find metric idx %d for %s",
					byTypeIdx, mdef.Name))
			return
		}
	}
	g.Add(val)
}

// Rec records a value (similar to what a Gauge would be)
func (pm *PrometheusMeter) Rec(key string, val float64) {
	pm.RecWL(key, val, nil)
}

// RecWL with labels that apply only to this record
func (pm *PrometheusMeter) RecWL(key string, val float64, labels map[string]string) {
	var err error
	mdef, idx := pm.metrics.catalog.Def(key)
	if mdef == nil {
		pm.log.WarnMsg("whop : metric def not found").Str("name", key).Str("mdef", fmt.Sprintf("%#v", pm.metrics.catalog)).Send()
		return
	}
	mtype := mdef.MetricType
	if mtype != MetricTypeHistogram {
		pm.log.Err(fmt.Errorf("bad Metric Operation"),
			fmt.Sprintf("cannot use Add with %s metric", key))
		return
	}
	byTypeIdx := pm.metrics.byTypeIdx[idx]
	var o prometheus.Observer
	if len(mdef.Labels) == 0 {
		o = pm.metrics.Histograms[byTypeIdx]
	} else {
		lbls := pm.fillLabels(mdef.Labels, labels)
		o, err = pm.metrics.LabeledHistograms[byTypeIdx].GetMetricWith(lbls)
		if err != nil {
			pm.log.Err(err, fmt.Sprintf("cannot get metric %s : %#v", key, pm.metrics))
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

func (pm *PrometheusMeter) fillLabels(labels []string, labelVals map[string]string) prometheus.Labels {
	promLabels := make(prometheus.Labels, len(labels))
	pm.dataMux.RLock()
	for _, l := range labels {
		v, ok := pm.data[l]
		if ok {
			promLabels[l] = v
		}
	}
	pm.dataMux.RUnlock()
	if labelVals == nil {
		return promLabels
	}
	for _, l := range labels {
		v, ok := labelVals[l]
		if ok {
			promLabels[l] = v
		}
	}
	return promLabels
}
