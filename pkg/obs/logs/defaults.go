package logs

import (
// "go.opentelemetry.io/otel/semconv/v1.27.0"
)

// This attributes can be used for traces and logs
const (
	// Only for Logs and Traces :
	AttrPath       string = "http.request.path"  // the full path
	AttrQuery      string = "http.request.query" // the query part of the uri
	AttrReqHeaders string = "http.request.headers"
	AttrReqID      string = "http.request.reqid"
	AttrRemoteIP   string = "http.request.remote_ip"

	// we can import the same name
	AttrReqDateTime string = "http.request.received" // time when the request was received
	AttrReqDuration string = "http.request.duration" // this is also a metric

	AttrRespSize string = "http.response.size"

	AttrErrors string = "errors"
)
