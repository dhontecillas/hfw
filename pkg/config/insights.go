package config

import (
	"fmt"

	"github.com/dhontecillas/hfw/pkg/obs"
	"github.com/dhontecillas/hfw/pkg/obs/logs"
	"github.com/dhontecillas/hfw/pkg/obs/metrics"
	"github.com/dhontecillas/hfw/pkg/obs/traces"
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
	TagDefs    []obs.TagDefinition
	MetricDefs metrics.Defs

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

func defaultMetricsConfig() metrics.Defs {
	requestLabels := []string{
		metrics.AttrRoute,
		metrics.AttrMethod,
	}

	responseLabels := []string{
		metrics.AttrStatus,
		metrics.AttrStatusGroup,
	}
	responseLabels = append(responseLabels, requestLabels...)

	dbconnErrorLabels := []string{
		metrics.AttrDBSQLAddress,
		metrics.AttrDBSQLDatasource,
	}
	dbconnErrorLabels = append(dbconnErrorLabels, requestLabels...)

	redisconnErrorLabels := []string{
		metrics.AttrDBRedisPool,
		metrics.AttrDBRedisAddress,
	}
	redisconnErrorLabels = append(redisconnErrorLabels, requestLabels...)

	distributionMetrics := map[string][]string{
		metrics.MetReqDuration:      responseLabels,
		metrics.MetHTTPResponseSize: responseLabels,
		metrics.MetDBQueryDuration:  responseLabels,
		metrics.MetDBConnError:      responseLabels,
		metrics.MetRedisConnError:   redisconnErrorLabels,
	}

	countMetrics := map[string][]string{
		metrics.MetReqCount:       responseLabels,
		metrics.MetReqTimeout:     responseLabels,
		metrics.MetDBConnError:    dbconnErrorLabels,
		metrics.MetRedisConnError: redisconnErrorLabels,
	}

	mDefs := make(metrics.Defs, 0, len(distributionMetrics))

	for dm, lbls := range distributionMetrics {
		mDefs = append(mDefs, metrics.Def{
			Name:       dm,
			MetricType: metrics.MetricTypeHistogram,
			Labels:     lbls,
		})
	}

	for cm, lbls := range countMetrics {
		mDefs = append(mDefs, metrics.Def{
			Name:       cm,
			MetricType: metrics.MetricTypeMonotonicCounter,
			Labels:     lbls,
		})
	}

	return mDefs
}

// ReadInsightsConfig reads the configuration for "routing" the
// tags to different subsystems (logs, metrics, traces). It also
// loads the configuration for logs / metrics / traces services.
func ReadInsightsConfig(confPrefix string) *InsightsConfig {
	conf := &InsightsConfig{
		TagDefs: []obs.TagDefinition{
			obs.TagDefinition{
				Name:    metrics.AttrStatus,
				TagType: obs.TagTypeI64,
				ToL:     true,
				ToM:     true,
				ToT:     true,
			},
			obs.TagDefinition{
				Name:    metrics.AttrRoute,
				TagType: obs.TagTypeStr,
				ToL:     true,
				ToM:     false,
				ToT:     true,
			},
			obs.TagDefinition{
				Name:    logs.AttrPath,
				TagType: obs.TagTypeStr,
				ToL:     true,
				ToM:     false,
				ToT:     true,
			},
			obs.TagDefinition{
				Name:    metrics.AttrMethod,
				TagType: obs.TagTypeStr,
				ToL:     true,
				ToM:     true,
				ToT:     true,
			},
			obs.TagDefinition{
				Name:    logs.AttrReqID,
				TagType: obs.TagTypeStr,
				ToL:     true,
				ToM:     false,
				ToT:     true,
			},
			obs.TagDefinition{
				Name:    logs.AttrRemoteIP,
				TagType: obs.TagTypeStr,
				ToL:     true,
				ToM:     false,
				ToT:     false,
			},
		},
	}
	conf.MetricDefs = defaultMetricsConfig()
	conf.loadPrometheusConfig(confPrefix)
	conf.loadGraylogConfig(confPrefix)
	conf.loadSentryConfig(confPrefix)
	return conf
}

// CreateInsightsBuilder creates a InsigheterBuilderFn and a flush
// function, based on the insConf configuration. It can also merge
// a list of additional metric definitions.
func CreateInsightsBuilder(insConf *InsightsConfig,
	metricDefs metrics.Defs) (obs.InsighterBuilderFn, func()) {

	if len(metricDefs) > 0 {
		mDefs := insConf.MetricDefs.Merge(metricDefs, false)
		insConf.MetricDefs = mDefs
	}

	logBuilder, logsFlushFn := newLoggerBuilder(insConf)
	l := logBuilder()

	meterBuilder, meterFlushFn := newMeterBuilder(l, insConf)

	nopTracerBuilder := traces.NewNopTracerBuilder()

	insB := obs.NewInsighterBuilder(insConf.TagDefs, logBuilder,
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
				metrics.AttrApp,
				metrics.AttrMethod,
				metrics.AttrRoute,
				logs.AttrPath,
				logs.AttrRemoteIP,
				logs.AttrReqID,
				metrics.AttrStatus,
				metrics.AttrStatusGroup,
			},
		}
		sentryBuilder, sentryFlush, err := logs.NewSentryBuilder(sentryConf)
		if err != nil {
			l.Err(err, "cannot create sentry logger builder")
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
			ServerPort:        strPort,
			ServerPath:        conf.PrometheusPath,
			MetricDefinitions: conf.MetricDefs,
			MetricsPrefix:     conf.PrometheusPrefix,
		}

		promMeterBuilder, err := metrics.NewPrometheusMeterBuilder(l, promConf)
		if err != nil {
			l.Err(err, "cannot create prometheus meter builder")
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
		l.Err(err, "cannot send metrics to multiple sinks, defaulting to first one")
		return enabledMeters[0], func() {}
	}
	return meterBuilder, func() {}
}
