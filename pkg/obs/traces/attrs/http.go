package traces

import (
	"go.opentelemetry.io/otel/semconv/v1.27.0"
)

// Signals and Attributes for HTTP Request Traces
const (
	AttrHTTPMethod      string = string(semconv.HTTPRequestMethodKey) // "method"
	AttrHTTPRoute       string = string(semconv.HTTPRouteKey)         // route
	AttrHTTPPath        string = "http.path"                          // path, with replaced params
	AttrHTTPStatus      string = string(semconv.HTTPResponseStatusCodeKey)
	AttrHTTPStatusGroup string = "http.response.status_group" // 2xx , 3xx, 4xx or 5xx , to reduce metric cardinality
	AttrHTTPQuery       string = "http.query"
	AttrHTTPRemoteIP    string = "http.remote_ip"
	AttrHTTPDuration    string = string(semconv.HTTPServerRequestDurationName)

	AttrHTTPRequestBodySize = string(semconv.HTTPRequestBodySizeKey) // http.request.size (including headers)
	AttrHTTPRequestSize     = string(semconv.HTTPRequestSizeKey)     // http.request.size (including headers)
	AttrHTTPRequestHeaders  = "http.request.headers"

	AttrHTTPResponseBodySize = string(semconv.HTTPResponseBodySizeKey) // http.request.size (including headers)
	AttrHTTPResponseSize     = string(semconv.HTTPResponseSizeKey)
	AttrHTTPResponseHeaders  = "http.response.headers"

	AttrHTTPErrors string = "http.errors"
)
