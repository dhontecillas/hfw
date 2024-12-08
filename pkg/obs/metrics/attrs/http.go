package attrs

import (
	"go.opentelemetry.io/otel/semconv/v1.27.0"

	obsattrs "github.com/dhontecillas/hfw/pkg/obs/attrs"
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

	AttrHTTPRequestID string = "http.request.id" //
	// -- HTTP metric signals

	// counter: with the attributes:
	// - method
	// - router
	MetHTTPServerRequestCount = "http.server.request.count"

	// histogram:
	MetHTTPServerRequestDuration = string(semconv.HTTPServerRequestDurationName)

	// histogram: http.server.request.body.size (including headers)
	MetHTTPServerRequestBodySize = string(semconv.HTTPServerRequestBodySizeName)

	// histogram: http.server.response.body.size (including headers)
	MetHTTPServerResponseBodySize = string(semconv.HTTPServerResponseBodySizeName)
)

var (
	AttrListHTTP = obsattrs.AttrDefinitionList{
		obsattrs.AttrDefinition{
			Name:        AttrHTTPMethod,
			StrAttrType: "str",
		},
		obsattrs.AttrDefinition{
			Name:        AttrHTTPRoute,
			StrAttrType: "str",
		},
		obsattrs.AttrDefinition{
			Name:        AttrHTTPStatus,
			StrAttrType: "i64",
		},
		obsattrs.AttrDefinition{
			Name:        AttrHTTPStatusGroup,
			StrAttrType: "i64",
		},
	}
)
