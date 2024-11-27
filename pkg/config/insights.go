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
	"github.com/spf13/viper"
)

const (
	confKeyPrometheusEnabled string = "prometheus.enabled"
	confKeyPrometheusPort    string = "prometheus.port"
	confKeyPrometheusPath    string = "prometheus.path"
	confKeyPrometheysPrefix  string = "prometheus.prefix"

	confKeySentryEnabled     string = "sentry.enabled"
	confKeySentryDSN         string = "sentry.dsn"
	confKeySentryEnvironment string = "env"

	confKeyGraylogEnabled string = "graylog.enabled"
	confKeyGraylogPort    string = "graylog.port"
	confKeyGraylogHost    string = "graylog.host"
	confKeyGraylogPrefix  string = "graylog.prefix"
)

// InsightsConfig holds the information for the metrics
type InsightsConfig struct {
	MetricDefs metrics.MetricDefinitionList

	PrometheusEnabled bool   // prometheus enabled or not
	PrometheusPort    int    // port where serving the metrics
	PrometheusPath    string // path to gather the metrics
	PrometheusPrefix  string // a prefix for all metrics

	GraylogEnabled   bool
	GraylogHost      string
	GraylogPort      int
	GraylogPrefix    string
	GraylogAddress   string
	GraylogConfError error

	SentryEnabled bool
	SentryDSN     string
	SentryEnv     string
}

func (ic *InsightsConfig) loadPrometheusConfig(confPrefix string) {
	promEnabled := viper.GetBool(confPrefix + confKeyPrometheusEnabled)
	if !promEnabled {
		return
	}
	ic.PrometheusEnabled = true
	promPort := viper.GetInt(confPrefix + confKeyPrometheusPort)
	if promPort < 1 && promPort > 65535 {
		promPort = 8090
	}
	ic.PrometheusPort = promPort

	promPath := viper.GetString(confPrefix + confKeyPrometheusPath)
	if len(promPath) == 0 {
		promPath = "/metrics"
	}
	ic.PrometheusPath = promPath

	ic.PrometheusPrefix = viper.GetString(confPrefix + confKeyPrometheysPrefix)
}

func (ic *InsightsConfig) loadGraylogConfig(confPrefix string) {
	ic.GraylogEnabled = viper.GetBool(confKeyGraylogEnabled)
	ic.GraylogHost = viper.GetString(confKeyGraylogHost)
	ic.GraylogPort = viper.GetInt(confKeyGraylogPort)
	ic.GraylogPrefix = viper.GetString(confKeyGraylogPrefix)

	if ic.GraylogEnabled {
		if ic.GraylogHost != "" && ic.GraylogPort > 0 {
			ic.GraylogAddress = fmt.Sprintf("%s:%d", ic.GraylogHost, ic.GraylogPort)
		} else {
			ic.GraylogConfError = fmt.Errorf(
				"graylog enabled and configuration not provided")
		}
	}
}

func (ic *InsightsConfig) loadSentryConfig(confPrefix string) {
	ic.SentryEnabled = viper.GetBool(confPrefix + confKeySentryEnabled)
	if ic.SentryEnabled {
		ic.SentryDSN = viper.GetString(confPrefix + confKeySentryDSN)
		ic.SentryEnv = viper.GetString(confPrefix + confKeySentryEnvironment)
	}
}

func defaultMetricsConfig() metrics.MetricDefinitionList {
	metricDefs := metricsdefaults.HTTPDefaultMetricDefinitions()
	return metricDefs
}

// ReadInsightsConfig reads the configuration for "routing" the
// tags to different subsystems (logs, metrics, traces). It also
// loads the configuration for logs / metrics / traces services.
func ReadInsightsConfig(confPrefix string) *InsightsConfig {
	conf := &InsightsConfig{
		MetricDefs: defaultMetricsConfig(),
	}
	conf.loadPrometheusConfig(confPrefix)
	conf.loadGraylogConfig(confPrefix)
	conf.loadSentryConfig(confPrefix)
	return conf
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

	insB := obs.NewInsighterBuilder(metricDefs, logBuilder,
		meterBuilder, nopTracerBuilder)

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
		GraylogHost:        conf.GraylogAddress,
		GraylogFieldPrefix: conf.GraylogPrefix,
	})
	if err != nil {
		panic(fmt.Sprintf("cannot build logger: %s", err.Error()))
	}
	flushers := []func(){logrusFlush}

	// we instantiate a logger for the rest of the initialization:
	l := logrusBuilder()
	loggerBuilders = append(loggerBuilders, logrusBuilder)

	if conf.SentryEnabled {
		sentryConf := &logs.SentryConf{
			Dsn:              conf.SentryDSN,
			AttachStacktrace: true,
			SampleRate:       1.0,
			// Release:       this should be a commit hash or something like that
			Environment:      conf.SentryEnv,
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
	if conf.PrometheusEnabled {
		strPort := fmt.Sprintf(":%d", conf.PrometheusPort)
		promConf := &metrics.PrometheusConfig{
			ServerPort:    strPort,
			ServerPath:    conf.PrometheusPath,
			MetricsPrefix: conf.PrometheusPrefix,
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
