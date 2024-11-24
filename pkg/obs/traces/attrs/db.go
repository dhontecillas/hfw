package traces

import (
	"go.opentelemetry.io/otel/semconv/v1.27.0"
)

// Signals and Attributes for Database Traces
const (

	// AttrDBQueryTextKey the full query statement 'SELECt * ...'
	// we should provide the query template, not the query with
	// the params replaced (to not leak sensitive data)
	// In a transactio,, all the queries should be concatenated with `;`
	AttrDBQueryText string = string(semconv.DBQueryTextKey)

	// AttrDBCollection name: makes only sense for a single query,
	// but not a full transaction
	AttrDBCollectionName = string(semconv.DBCollectionNameKey)

	// AttrDBNamespace the fully qualified name of the connection to the db:
	// host:port:dbname:namespace
	AttrDBNamespace = string(semconv.DBNamespaceKey)

	// AttrDBOperationName should be a SELECT, HMSET or single type operation by the standard
	// (not bery useful if we include the DBQuery Text).
	AttrDBOperationName = string(semconv.DBOperationNameKey)

	// if is a 'sql', 'redis', or any type of database system
	AttrDBSystem = string(semconv.DBSystemKey)
)
