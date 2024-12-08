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

func (m *NopMeter) Clone() Meter {
	return m
}

// Inc increases an integer value
func (m *NopMeter) Inc(key string) {
}

// IncWL increases and integer value adding labels to this records
func (m *NopMeter) IncWL(key string, labels map[string]interface{}) {
}

// Dec decreases an integer value
func (m *NopMeter) Dec(key string) {
}

// DecWL decreases an integer value adding labels to this record
func (m *NopMeter) DecWL(key string, labels map[string]interface{}) {
}

// Add adds an int value
func (m *NopMeter) Add(key string, val int64) {
}

// AddWL with labels that apply only to this record
func (m *NopMeter) AddWL(key string, val int64, labels map[string]interface{}) {
}

// Rec records a value (similar to what a Gauge would be)
func (m *NopMeter) Rec(key string, val float64) {
}

// RecWL with labels that apply only to this record
func (m *NopMeter) RecWL(key string, val float64, labels map[string]interface{}) {
}

// Str sets a label value for all metrics that have defined it
func (m *NopMeter) Str(key string, val string) {
}

// Attrs sets several values for all metrics that have defined it
func (m *NopMeter) SetAttrs(attrMap map[string]interface{}) {
}

// NewNopMeterBuilder returns a func
func NewNopMeterBuilder() (MeterBuilderFn, error) {
	return func(log logs.Logger) Meter {
		return NewNopMeter()
	}, nil
}
