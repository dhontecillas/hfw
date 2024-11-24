package ginfw

import (
	"net/http"
	"strings"
	"time"

	"github.com/dhontecillas/hfw/pkg/ids"
	"github.com/gin-gonic/gin"

	"github.com/dhontecillas/hfw/pkg/obs/logs"
	"github.com/dhontecillas/hfw/pkg/obs/metrics"
)

func headersToString(headers http.Header) string {
	var hb strings.Builder
	// TODO: we want the headers sorted by name
	for k, v := range headers {
		hb.WriteString(k)
		hb.WriteString(":[")
		for idx, vv := range v {
			if idx > 0 {
				hb.WriteString(",")
			}
			hb.WriteString(vv)
		}
		hb.WriteString("] ")
	}
	return hb.String()
}

func statusGroup(status int) string {
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

// ObsMiddleware creates a middlewware that attaches the
// ExternalServices instance and insighters instance with
// some fields already set, like the method and path for
// the request, the client ip for the logs, the request time
// as well as a unique request ID. It will also report the
// duration and size of the response.
func ObsMiddleware() gin.HandlerFunc {
	idGen := ids.NewIDGenerator()

	return func(c *gin.Context) {
		deps := ExtServices(c)
		ins := deps.Ins

		// set some shared tags for logs, metrics and traces
		ins.Str(metrics.AttrMethod, c.Request.Method)
		ins.Str(metrics.AttrRoute, c.FullPath())

		ins.L.Str(logs.AttrPath, c.Request.URL.Path)
		ins.T.Str(logs.AttrPath, c.Request.URL.Path)
		ins.L.Str(logs.AttrQuery, c.Request.URL.RawQuery)
		ins.T.Str(logs.AttrQuery, c.Request.URL.Path)

		strHeaders := headersToString(c.Request.Header)
		ins.L.Str(logs.AttrReqHeaders, strHeaders)
		ins.T.Str(logs.AttrReqHeaders, strHeaders)

		reqID := "UNKNOWN"
		if id, err := idGen.New(); err == nil {
			reqID = id.ToUUID()
		}
		ins.Str(logs.AttrReqID, reqID)

		// set shared tags for all logs
		ins.L.Str(logs.AttrRemoteIP, c.ClientIP())

		startTime := time.Now()
		ins.L.Str(logs.AttrReqDateTime, startTime.Format(time.RFC3339))

		c.Next()

		// tag the metrics with the status code
		status := c.Writer.Status()
		statusG := statusGroup(status)

		ins.I64(metrics.AttrStatus, int64(status))
		ins.Str(metrics.AttrStatusGroup, statusG)

		duration := time.Since(startTime).Milliseconds()
		respSize := c.Writer.Size()

		// TODO: we need to record the size of the incoming payload

		// meter the duration of the endpoint
		ins.M.Rec(metrics.MetReqDuration, float64(duration))
		ins.M.Rec(metrics.MetHTTPResponseBodySize, float64(respSize))

		ins.L.I64(logs.AttrReqDuration, duration)
		ins.L.I64(logs.AttrRespSize, int64(respSize))

		m := ins.L.InfoMsg("Request")

		errs := c.Errors.ByType(gin.ErrorTypePrivate).String()
		if len(errs) > 0 {
			m.Str(logs.AttrErrors, errs)
		}
		m.Send()
	}
}
