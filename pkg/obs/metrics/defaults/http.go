package defaults

import (
	"github.com/dhontecillas/hfw/pkg/obs/metrics"
	metattrs "github.com/dhontecillas/hfw/pkg/obs/metrics/attrs"
)

func HTTPDefaultMetricDefinitions() metrics.MetricDefinitionList {
	return metrics.MetricDefinitionList{
		&metrics.MetricDefinition{
			Name:       metattrs.MetHTTPServerRequestCount,
			Units:      "",
			MetricType: metrics.MetricTypeMonotonicCounter,
			Attributes: metattrs.AttrListHTTP,
		},
		&metrics.MetricDefinition{
			Name:       metattrs.MetHTTPServerRequestDuration,
			Units:      "s",
			MetricType: metrics.MetricTypeHistogram,
			Attributes: metattrs.AttrListHTTP,
		},
		&metrics.MetricDefinition{
			Name:       metattrs.MetHTTPServerRequestBodySize,
			Units:      "b",
			MetricType: metrics.MetricTypeHistogram,
			Attributes: metattrs.AttrListHTTP,
		},
		&metrics.MetricDefinition{
			Name:       metattrs.MetHTTPServerResponseBodySize,
			Units:      "b",
			MetricType: metrics.MetricTypeHistogram,
			Attributes: metattrs.AttrListHTTP,
		},
	}
}
