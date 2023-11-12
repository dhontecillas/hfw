package ginfw

import (
	"strings"
	"time"

	"github.com/dhontecillas/hfw/pkg/ids"
	"github.com/gin-gonic/gin"
)

// Constants for configuration parameters related to logs / metrics / traces
const (
	// tags shared across all insights ?
	LabelInsApp string = "app"

	LabelInsMethod   string = "method"
	LabelInsPath     string = "path"
	LabelInsQuery    string = "query"
	LabelInsHeaders  string = "headers"
	LabelInsReqID    string = "reqid"
	LabelInsRemoteIP string = "ip"
	LabelInsStatus   string = "status"

	// Metric definitions for database

	// MetDBConnError keeps track of the number of db conn. errors
	MetDBConnError string = "db.connerror"
	MetDBDuration  string = "db.duration"

	LabelMetDBAddress    string = "address"
	LabelMetDBDatasource string = "datasource"

	// Metric definitions for Redis
	MetRedisConnError string = "redis.connerror"

	LabelMetRedisPool    string = "pool"
	LabelMetRedisAddress string = "address"

	// Metric definitions for requests
	MetReqCount    string = "request.count"
	MetReqDuration string = "request.duration"
	MetReqTimeout  string = "request.timeout"
	MetReqSize     string = "request.size"

	// LogTime
	LabelLogTime     string = "reqtime"
	LabelLogDuration string = "duration"
	LabelLogSize     string = "size"
	LabelLogErrors   string = "errors"
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

		// set some shared tags for logs, metrics and traces
		ins.Str(LabelInsMethod, c.Request.Method)
		ins.Str(LabelInsPath, c.Request.URL.Path)
		ins.L.Str(LabelInsQuery, c.Request.URL.RawQuery)

		var hb strings.Builder
		for k, v := range c.Request.Header {
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
		ins.L.Str(LabelInsHeaders, hb.String())

		reqID := "UNKNOWN"
		if id, err := idGen.New(); err == nil {
			reqID = id.ToUUID()
		}
		ins.Str(LabelInsReqID, reqID)

		// set shared tags for all logs
		ins.L.Str(LabelInsRemoteIP, c.ClientIP())
		startTime := time.Now()
		ins.L.Str(LabelLogTime, startTime.Format(time.RFC3339))

		c.Next()

		// tag the metrics with the status code
		status := c.Writer.Status()
		ins.I64(LabelInsStatus, int64(status))

		duration := time.Since(startTime).Milliseconds()
		size := c.Writer.Size()

		// meter the duration of the endpoint
		ins.M.Rec(MetReqDuration, float64(duration))
		ins.M.Rec(MetReqSize, float64(size))

		ins.L.I64(LabelLogDuration, duration)
		ins.L.I64(LabelLogSize, int64(size))

		m := ins.L.InfoMsg("Request")
		errs := c.Errors.ByType(gin.ErrorTypePrivate).String()
		if len(errs) > 0 {
			m.Str(LabelLogErrors, errs)
		}
		m.Send()
	}
}
