package ginfw

import (
	"time"

	"github.com/dhontecillas/hfw/pkg/ids"
	"github.com/gin-gonic/gin"

	"github.com/dhontecillas/hfw/pkg/obs/httpobs"
	metricsattrs "github.com/dhontecillas/hfw/pkg/obs/metrics/attrs"
	tracesattrs "github.com/dhontecillas/hfw/pkg/obs/traces/attrs"
)

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
		req := c.Request

		reqAttrs, _ := httpobs.HTTPRequestAttrs(req, c.FullPath())
		ins.SetAttrs(reqAttrs)
		ins.Str(tracesattrs.AttrHTTPRemoteIP, c.ClientIP())
		reqID := "UNKNOWN"
		if id, err := idGen.New(); err == nil {
			reqID = id.ToUUID()
		}
		ins.Str(metricsattrs.AttrHTTPRequestID, reqID)

		span := ins.T.Start(req.Context(), "http_request", reqAttrs)
		defer span.End()

		// set some shared tags for logs, metrics and traces

		// set shared tags for all logs

		startTime := time.Now()
		ins.L.Str("time", startTime.Format(time.RFC3339))

		c.Next()

		// tag the metrics with the status code
		status := c.Writer.Status()
		statusG := httpobs.HTTPStatusGroup(status)

		ins.I64(metricsattrs.AttrHTTPStatus, int64(status))
		ins.Str(metricsattrs.AttrHTTPStatusGroup, statusG)

		duration := time.Since(startTime).Milliseconds()
		respSize := c.Writer.Size()

		// TODO: we need to record the size of the incoming payload

		// meter the duration of the endpoint
		ins.M.Rec(metricsattrs.MetHTTPServerRequestDuration, float64(duration))
		ins.M.Rec(metricsattrs.MetHTTPServerResponseBodySize, float64(respSize))

		ins.T.I64(tracesattrs.AttrHTTPDuration, duration)
		ins.T.I64(tracesattrs.AttrHTTPResponseBodySize, int64(respSize))
		ins.L.I64(tracesattrs.AttrHTTPDuration, duration)
		ins.L.I64(tracesattrs.AttrHTTPResponseBodySize, int64(respSize))

		errs := c.Errors.ByType(gin.ErrorTypePrivate).String()
		ins.L.Info("request", map[string]interface{}{
			"errors": errs,
		})
	}
}
