package metrics

import (
	"github.com/dhontecillas/hfw/pkg/obs/logs"
)

// MockMeter is an implementation of a meter used to
// check that we are sending
type MockMeter struct {
	Incs     []string
	Decs     []string
	Recs     []string
	RecsVals []float64
	Adds     []string
	AddsVals []int64

	Strs map[string]string
}

// NewMockMeter creates a new Mock meter.
func NewMockMeter() *MockMeter {
	return &MockMeter{
		Incs:     []string{},
		Decs:     []string{},
		Recs:     []string{},
		RecsVals: []float64{},
		Adds:     []string{},
		AddsVals: []int64{},
		Strs:     map[string]string{},
	}
}

var _ Meter = (*MockMeter)(nil)

// Inc increases an integer value
func (m *MockMeter) Inc(key string) {
	m.Incs = append(m.Incs, key)
}

// IncWL increases and integer value adding labels to this records
func (m *MockMeter) IncWL(key string, labels map[string]string) {
	m.Incs = append(m.Incs, key)
}

// Dec decreases an integer value
func (m *MockMeter) Dec(key string) {
	m.Decs = append(m.Decs, key)
}

// DecWL decreases an integer value adding labels to this record
func (m *MockMeter) DecWL(key string, labels map[string]string) {
	m.Decs = append(m.Decs, key)
}

// Add adds an int value
func (m *MockMeter) Add(key string, val int64) {
	m.Adds = append(m.Adds, key)
	m.AddsVals = append(m.AddsVals, val)
}

// AddWL with labels that apply only to this record
func (m *MockMeter) AddWL(key string, val int64, labels map[string]string) {
	m.Adds = append(m.Adds, key)
	m.AddsVals = append(m.AddsVals, val)
}

// Rec records a value (similar to what a Gauge would be)
func (m *MockMeter) Rec(key string, val float64) {
	m.Recs = append(m.Recs, key)
	m.RecsVals = append(m.RecsVals, val)
}

// RecWL with labels that apply only to this record
func (m *MockMeter) RecWL(key string, val float64, labels map[string]string) {
	m.Recs = append(m.Recs, key)
	m.RecsVals = append(m.RecsVals, val)
}

// Str sets a label value for all metrics that have defined it
func (m *MockMeter) Str(key string, val string) {
	m.Strs[key] = val
}

// NewMockMeterBuilder returns a function to create Mock meters.
func NewMockMeterBuilder() (MeterBuilderFn, error) {
	return func(log logs.Logger) Meter {
		return NewMockMeter()
	}, nil
}
