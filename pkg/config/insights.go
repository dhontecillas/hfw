package config

import (
	"fmt"

	"github.com/dhontecillas/hfw/pkg/obs"
	"github.com/dhontecillas/hfw/pkg/obs/logs"
	"github.com/dhontecillas/hfw/pkg/obs/metrics"
	metricattrs "github.com/dhontecillas/hfw/pkg/obs/metrics/attrs"
	metricsdefaults "github.com/dhontecillas/hfw/pkg/obs/metrics/defaults"
	"github.com/dhontecillas/hfw/pkg/obs/traces"
	tracesattrs "github.com/dhontecillas/hfw/pkg/obs/traces/attrs"
)

const ()

var (
	ErrBadPortNumber = fmt.Errorf("bad port number")
	ErrMissingHost   = fmt.Errorf("missing host")
	ErrMissingDSN    = fmt.Errorf("missing DSN")
)

type InsightsPrometheusConfig struct {
	Enabled bool   `json:"enabled"` // prometheus enabled or not
	Port    int    `json:"port"`    // port where serving the metrics
	Path    string `json:"path"`    // path to gather the metrics
	Prefix  string `json:"prefix"`  // a prefix for all metrics
}

func (c *InsightsPrometheusConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Port < 0 || c.Port > 65536 {
		return ErrBadPortNumber
	}
	if c.Port == 0 {
		c.Port = 8090 // we use 8090 as default port
	}
	if c.Path == "" {
		c.Path = "/metrics"
	}
	return nil
}

type InsightsGraylogConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Prefix  string `json:"prefix"`
	Address string // `json:"address"`
}

func (c *InsightsGraylogConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Host == "" {
		return ErrMissingHost
	}
	if c.Port < 0 || c.Port > 65536 {
		return ErrBadPortNumber
	}
	if c.Port == 0 {
		c.Port = 9000
	}
	c.Address = fmt.Sprintf("%s:%d", c.Host, c.Port)
	return nil
}

type InsightsSentryConfig struct {
	Enabled bool   `json:"enabled"`
	DSN     string `json:"dsn"`
	Env     string `json:"env"`
}

func (c *InsightsSentryConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if len(c.DSN) == 0 {
		return ErrMissingDSN
	}
	return nil
}

// InsightsConfig holds the information for the metrics
type InsightsConfig struct {
	MetricDefs metrics.MetricDefinitionList

	Prometheus InsightsPrometheusConfig `json:"prometheus"`
	Graylog    InsightsGraylogConfig    `json:"graylog"`
	Sentry     InsightsSentryConfig     `json:"sentry"`
}

func (c *InsightsConfig) Validate() error {
	if err := c.Prometheus.Validate(); err != nil {
		return err
	}
	if err := c.Graylog.Validate(); err != nil {
		return err
	}
	if err := c.Sentry.Validate(); err != nil {
		return err
	}
	return nil
}

func defaultMetricsConfig() metrics.MetricDefinitionList {
	metricDefs := metricsdefaults.HTTPDefaultMetricDefinitions()
	return metricDefs
}

// ReadInsightsConfig reads the configuration for "routing" the
// tags to different subsystems (logs, metrics, traces). It also
// loads the configuration for logs / metrics / traces services.
func ReadInsightsConfig(cldr ConfLoader) *InsightsConfig {
	var err error
	conf := InsightsConfig{}
	defer func() {
		conf.MetricDefs = defaultMetricsConfig()
	}()

	cldr, err = cldr.Section([]string{"insights"})
	if err != nil {
		return &conf
	}
	if err := cldr.Parse(&conf); err != nil {
		conf.MetricDefs = defaultMetricsConfig()
	}
	conf.Validate()
	return &conf
}

// CreateInsightsBuilder creates a InsigheterBuilderFn and a flush
// function, based on the insConf configuration. It can also merge
// a list of additional metric definitions.
func CreateInsightsBuilder(insConf *InsightsConfig,
	metricDefs metrics.MetricDefinitionList) (obs.InsighterBuilderFn, func()) {

	if len(metricDefs) > 0 {
		mDefs := insConf.MetricDefs.Merge(metricDefs, false)
		insConf.MetricDefs = mDefs
	}

	logBuilder, logsFlushFn := newLoggerBuilder(insConf)
	l := logBuilder()

	meterBuilder, meterFlushFn := newMeterBuilder(l, insConf)

	nopTracerBuilder := traces.NewNopTracerBuilder()

	insB := obs.NewInsighterBuilder(logBuilder, meterBuilder, nopTracerBuilder)

	flushFn := multiFlushFn(logsFlushFn, meterFlushFn)
	return insB, flushFn
}

func multiFlushFn(fns ...func()) func() {
	return func() {
		for _, fn := range fns {
			fn()
		}
	}
}

func newLoggerBuilder(conf *InsightsConfig) (logs.LoggerBuilderFn, func()) {

	loggerBuilders := []logs.LoggerBuilderFn{}
	logrusBuilder, logrusFlush, err := logs.NewLogrusBuilder(&logs.LogrusConf{
		OutFileName:        "",
		GraylogHost:        conf.Graylog.Address,
		GraylogFieldPrefix: conf.Graylog.Prefix,
	})
	if err != nil {
		panic(fmt.Sprintf("cannot build logger: %s", err.Error()))
	}
	flushers := []func(){logrusFlush}

	// we instantiate a logger for the rest of the initialization:
	l := logrusBuilder()
	loggerBuilders = append(loggerBuilders, logrusBuilder)

	if conf.Sentry.Enabled {
		sentryConf := &logs.SentryConf{
			Dsn:              conf.Sentry.DSN,
			AttachStacktrace: true,
			SampleRate:       1.0,
			// Release:       this should be a commit hash or something like that
			Environment:      conf.Sentry.Env,
			FlushTimeoutSecs: 2,
			LevelThreshold:   "warning",
			AllowedTags: []string{
				metricattrs.AttrApp,
				metricattrs.AttrHTTPMethod,
				metricattrs.AttrHTTPRoute,
				tracesattrs.AttrHTTPPath,
				tracesattrs.AttrHTTPRemoteIP,
				"req_id",
				metricattrs.AttrHTTPStatus,
				metricattrs.AttrHTTPStatusGroup,
			},
		}
		sentryBuilder, sentryFlush, err := logs.NewSentryBuilder(sentryConf)
		if err != nil {
			l.Err(err, "cannot create sentry logger builder", nil)
		} else {
			loggerBuilders = append(loggerBuilders, sentryBuilder)
			flushers = append(flushers, sentryFlush)
		}
	}

	loggerBuilder := loggerBuilders[0]
	if len(loggerBuilders) > 1 {
		loggerBuilder = logs.NewMultiLoggerBuilder(loggerBuilders...)
	}

	return loggerBuilder, multiFlushFn(flushers...)
}

func newMeterBuilder(l logs.Logger, conf *InsightsConfig) (metrics.MeterBuilderFn, func()) {
	var enabledMeters []metrics.MeterBuilderFn
	if conf.Prometheus.Enabled {
		strPort := fmt.Sprintf(":%d", conf.Prometheus.Port)
		promConf := &metrics.PrometheusConfig{
			ServerPort:    strPort,
			ServerPath:    conf.Prometheus.Path,
			MetricsPrefix: conf.Prometheus.Prefix,
		}

		promMeterBuilder, err := metrics.NewPrometheusMeterBuilder(l, promConf,
			conf.MetricDefs)
		if err != nil {
			l.Err(err, "cannot create prometheus meter builder", nil)
		} else {
			enabledMeters = append(enabledMeters, promMeterBuilder)
		}
		metrics.Serve(promConf)
	}

	if len(enabledMeters) == 0 {
		nopMeterBuilder, _ := metrics.NewNopMeterBuilder()
		enabledMeters = append(enabledMeters, nopMeterBuilder)
	}

	if len(enabledMeters) == 1 {
		return enabledMeters[0], func() {}
	}
	meterBuilder, err := metrics.NewMultiMeterBuilder(l, enabledMeters...)
	if err != nil {
		l.Err(err, "cannot send metrics to multiple sinks, defaulting to first one", nil)
		return enabledMeters[0], func() {}
	}
	return meterBuilder, func() {}
}
