package httpobs

import (
	"fmt"
	"net/http"

	metricattrs "github.com/dhontecillas/hfw/pkg/obs/metrics/attrs"
	traceattrs "github.com/dhontecillas/hfw/pkg/obs/traces/attrs"
)

// TODO: how to make fields private ?
// auth headers ? cookies ?
//
// ExtractTelemetryRequestAndFields creates a new request with fields
// that should be kept private removed, and a list of tags that
// can be fed to the Insighter interface
func ExtractTelemetryFields(req *http.Request) (map[string]interface{}, error) {
	if req == nil {
		return nil, fmt.Errorf("nil pointer")
	}

	// TODO: add headers, and cookies, current timestamp
	return map[string]interface{}{
		metricattrs.AttrHTTPMethod: req.Method,
		traceattrs.AttrHTTPPath:    req.URL.Path,
	}, nil
}

type ObsResponseWriter struct {
}

type ObsBodyReader struct {
}

type ObsBodyWriter struct {
}
