package metrics

import (
	"github.com/dhontecillas/hfw/pkg/obs/logs"
)

// NopMeter is a meter that does nothing
type NopMeter struct {
}

// NewNopMeter creates a NopMeter
func NewNopMeter() *NopMeter {
	return &NopMeter{}
}

var _ Meter = (*NopMeter)(nil)

// Inc increases an integer value
func (m *NopMeter) Inc(key string) {
}

// IncWL increases and integer value adding labels to this records
func (m *NopMeter) IncWL(key string, labels map[string]string) {
}

// Dec decreases an integer value
func (m *NopMeter) Dec(key string) {
}

// DecWL decreases an integer value adding labels to this record
func (m *NopMeter) DecWL(key string, labels map[string]string) {
}

// Add adds an int value
func (m *NopMeter) Add(key string, val float64) {
}

// AddWL with labels that apply only to this record
func (m *NopMeter) AddWL(key string, val float64, labels map[string]string) {
}

// Rec records a value (similar to what a Gauge would be)
func (m *NopMeter) Rec(key string, val float64) {
}

// RecWL with labels that apply only to this record
func (m *NopMeter) RecWL(key string, val float64, labels map[string]string) {
}

// Str sets a label value for all metrics that have defined it
func (m *NopMeter) Str(key string, val string) {
}

// NewNopMeterBuilder returns a func
func NewNopMeterBuilder() (MeterBuilderFn, error) {
	return func(log logs.Logger) Meter {
		return NewNopMeter()
	}, nil
}
