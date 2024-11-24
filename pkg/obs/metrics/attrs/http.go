package metrics

import (
	"go.opentelemetry.io/otel/semconv/v1.27.0"
)

// Attributes that can appear in http requests
const (
	// -- HTTP metric attributes

	// request metric attributes
	AttrHTTPMethod string = string(semconv.HTTPRequestMethodKey) // "method"
	AttrHTTPRoute  string = string(semconv.HTTPRouteKey)         // route

	// response metric attributes
	AttrHTTPStatus      string = string(semconv.HTTPResponseStatusCodeKey)
	AttrHTTPStatusGroup string = "http.response.status_group" // 2xx , 3xx, 4xx or 5xx , to reduce metric cardinality

	// -- HTTP metric signals

	// counter: with the attributes:
	// - method
	// - router
	MetHTTPServerRequestCount = "http.server.request.count"

	// histogram:
	MetHTTPServerRequestDuration = string(semconv.HTTPServerRequestDurationName)

	// histogram: http.server.request.body.size (including headers)
	MetHTTPRequestBodySize = string(semconv.HTTPServerRequestBodySizeName)

	// histogram: http.server.response.body.size (including headers)
	MetHTTPResponseBodySize = string(semconv.HTTPServerResponseBodySizeName)
)
