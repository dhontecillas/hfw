package attrs

import (
	"go.opentelemetry.io/otel/semconv/v1.27.0"
)

// Constants for metric signals, and its attributes

// Attributes shared across all signal metrics
const (
	AttrApp string = "app"
)

// Signals and attributes for database metrics
const (
	AttrDBSystem    string = string(semconv.DBSystemKey)
	AtrrDBNamespace string = string(semconv.DBNamespaceKey)

	AttrDBErrorType string = string(semconv.ErrorTypeKey)

	// histogram: (the
	MetDBClientOperationDuration string = string(semconv.DBClientOperationDurationName)

	// counter with error type attribute (AttrDBErrorType), with limited
	// types to:
	// - QUERY error
	// - CONN error
	// - TIMEOUT error
	MetDBClientConnectionTimouts string = "db.client.error"
)

const (

	// MetDBConnError keeps track of the number of db conn. errors
	MetDBConnError     string = "db.client.sql.connerror"
	MetDBQueryDuration string = "db.client.sql.query.duration"

	// Metric definitions for Redis
	MetRedisConnError     string = "db.client.redis.connerror"
	MetRedisQueryDuration string = "db.client.redis.query.duration"

	// Metric definitions for requests
	MetReqCount    string = "http.request.count"
	MetReqDuration string = "http.request.duration"
	MetReqTimeout  string = "http.request.timeout"
)
