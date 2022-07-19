package metrics

import (
	"github.com/dhontecillas/hfw/pkg/obs/logs"
)

/*
type StatsD interface {
	Inc()
	Dec()
	Count()
	Gauge()

	// Check how these two works:
	Histogram(string, float64, []string, float64) error
	Distribution(string, float64, []string, float64) error
}

type OpenTelemetryMetrics interface {
	Inc(key string)
	Dec(key string)
	Add(key string, val string)
}
*/

// Meter is the interface that abstracts the underlying metrics system.
// All counters used work on Float64 values
type Meter interface {
	// Inc increases an integer value
	Inc(key string)
	// Dec decreases an integer value
	Dec(key string)
	// Add adds an int value
	Add(key string, val float64)
	// Rec records a value (similar to what a Gauge would be)
	Rec(key string, val float64)

	// Str sets a label value for all metrics that have defined it
	Str(key string, val string)

	// IncWL increases and integer value adding labels to this records
	IncWL(key string, labels map[string]string)
	// DecWL decreases an integer value adding labels to this record
	DecWL(key string, labels map[string]string)
	// AddWL with labels that apply only to this record
	AddWL(key string, val float64, labels map[string]string)
	// RecWL with labels that apply only to this record
	RecWL(key string, val float64, labels map[string]string)
}

// MeterBuilderFn defines the type of a builder function for Meters
type MeterBuilderFn func(log logs.Logger) Meter
