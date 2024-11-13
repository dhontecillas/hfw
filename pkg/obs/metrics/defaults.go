package metrics

import (
	"go.opentelemetry.io/otel/semconv/v1.27.0"
)

// Constants for metric signals, and its attributes

const (
	// tags shared across all insights ?
	AttrApp string = "app"

	// request metric attributes
	AttrMethod string = string(semconv.HTTPRequestMethodKey) // "method"
	AttrRoute  string = string(semconv.HTTPRouteKey)         // router

	// response metric attributes
	AttrStatus      string = string(semconv.HTTPResponseStatusCodeKey)
	AttrStatusGroup string = "http.response.status_group" // 2xx , 3xx, 4xx or 5xx , to reduce metric cardinality

	AttrDBSQLAddress    string = "db.sql.address"
	AttrDBSQLDatasource string = "db.sql.datasource"

	AttrDBRedisAddress string = "db.redis.address"
	AttrDBRedisPool    string = "db.redis.pool"
)

// Signals for metrics (or name of metrics)
const (

	// request metrics:
	MetHTTPRequestBodySize = string(semconv.HTTPRequestBodySizeKey) // http.request.size (including headers)
	MetHTTPRequestSize     = string(semconv.HTTPRequestSizeKey)     // http.request.size (including headers)

	// response metrics:
	MetHTTPResponseBodySize = string(semconv.HTTPResponseBodySizeKey) // http.request.size (including headers)
	MetHTTPResponseSize     = string(semconv.HTTPResponseSizeKey)

	// MetDBConnError keeps track of the number of db conn. errors
	MetDBConnError     string = "db.sql.connerror"
	MetDBQueryDuration string = "db.sql.query.duration"

	// Metric definitions for Redis
	MetRedisConnError     string = "db.redis.connerror"
	MetRedisQueryDuration string = "db.redis.query.duration"

	// Metric definitions for requests
	MetReqCount    string = "http.request.count"
	MetReqDuration string = "http.request.duration"
	MetReqTimeout  string = "http.request.timeout"
)
