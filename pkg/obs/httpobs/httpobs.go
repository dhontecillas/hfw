package httpobs

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dhontecillas/hfw/pkg/obs"
	metricattrs "github.com/dhontecillas/hfw/pkg/obs/metrics/attrs"
	traceattrs "github.com/dhontecillas/hfw/pkg/obs/traces/attrs"
)

// TODO: how to make fields private ?
// auth headers ? cookies ?
//
// ExtractTelemetryRequestAndFields creates a new request with fields
// that should be kept private removed, and a list of tags that
// can be fed to the Insighter interface
func HTTPRequestAttrs(req *http.Request, route string) (map[string]interface{}, error) {
	if req == nil {
		return nil, fmt.Errorf("nil pointer")
	}

	// TODO: add headers, and cookies, current timestamp
	return map[string]interface{}{
		metricattrs.AttrHTTPMethod:  req.Method,
		metricattrs.AttrHTTPRoute:   route,
		traceattrs.AttrHTTPPath:     req.URL.Path,
		traceattrs.AttrHTTPQuery:    req.URL.Query(),
		traceattrs.AttrHTTPRemoteIP: req.RemoteAddr,
		// TODO: implement a body reader that counts the number of bytes
		// traceattrs.AttrHTTPRequestBodySize:
	}, nil
}

func HTTPStatusGroup(status int) string {
	if status < 200 {
		return "1xx"
	}
	if status < 300 {
		return "2xx"
	}
	if status < 400 {
		return "3xx"
	}
	if status < 500 {
		return "4xx"
	}
	return "5xx"
}

func HTTPResponseAttrs(resp *http.Response) (map[string]interface{}, error) {
	if resp == nil {
		return nil, fmt.Errorf("nil pointer")
	}
	return map[string]interface{}{
		metricattrs.AttrHTTPStatus:      resp.StatusCode,
		metricattrs.AttrHTTPStatusGroup: HTTPStatusGroup(resp.StatusCode),
	}, nil
}

func NewObsHTTPHandler(insBuilder obs.InsighterBuilderFn, next http.HandlerFunc) http.HandlerFunc {
	ins := insBuilder()
	return func(w http.ResponseWriter, req *http.Request) {
		reqAttrs, _ := HTTPRequestAttrs(req, "")
		span := ins.T.Start(req.Context(), "http_request", reqAttrs)
		ins.M.SetAttrs(reqAttrs)

		started := time.Now()
		next(w, req)

		// TODO: we need a way to extract the status code
		duration := time.Now().Sub(started)
		secs := duration.Seconds()

		fmt.Printf("DBG: Secs: %f\n", secs)
		ins.M.Rec(metricattrs.MetHTTPServerRequestDuration, secs)
		span.F64(traceattrs.AttrHTTPDuration, secs)
		span.End()
	}
}
