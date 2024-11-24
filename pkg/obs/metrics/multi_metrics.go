package metrics

import (
	"github.com/dhontecillas/hfw/pkg/obs/logs"
)

// MultiMeter is a meter that allows to send metrics
// to several other meters.
type MultiMeter struct {
	wrapped []Meter
}

// NewMultiMeter creates a MultiMeter
func NewMultiMeter(wrapped ...Meter) *MultiMeter {
	m := &MultiMeter{}
	m.wrapped = make([]Meter, len(wrapped))
	copy(m.wrapped, wrapped)
	return m
}

var _ Meter = (*MultiMeter)(nil)

// Inc increases an integer value
func (m *MultiMeter) Inc(key string) {
	for _, w := range m.wrapped {
		w.Inc(key)
	}
}

// IncWL increases and integer value adding labels to this records
func (m *MultiMeter) IncWL(key string, labels map[string]string) {
	for _, w := range m.wrapped {
		w.IncWL(key, labels)
	}
}

// Dec decreases an integer value
func (m *MultiMeter) Dec(key string) {
	for _, w := range m.wrapped {
		w.Dec(key)
	}
}

// DecWL decreases an integer value adding labels to this record
func (m *MultiMeter) DecWL(key string, labels map[string]string) {
	for _, w := range m.wrapped {
		w.DecWL(key, labels)
	}
}

// Add adds an int value
func (m *MultiMeter) Add(key string, val int64) {
	for _, w := range m.wrapped {
		w.Add(key, val)
	}
}

// AddWL with labels that apply only to this record
func (m *MultiMeter) AddWL(key string, val int64, labels map[string]string) {
	for _, w := range m.wrapped {
		w.AddWL(key, val, labels)
	}
}

// Rec records a value (similar to what a Gauge would be)
func (m *MultiMeter) Rec(key string, val float64) {
	for _, w := range m.wrapped {
		w.Rec(key, val)
	}
}

// RecWL with labels that apply only to this record
func (m *MultiMeter) RecWL(key string, val float64, labels map[string]string) {
	for _, w := range m.wrapped {
		w.RecWL(key, val, labels)
	}
}

// Str sets a label value for all metrics that have defined it
func (m *MultiMeter) Str(key string, val string) {
	for _, w := range m.wrapped {
		w.Str(key, val)
	}
}

// NewMultiMeterBuilder crates a function to build NopMeters
func NewMultiMeterBuilder(log logs.Logger, builders ...MeterBuilderFn) (MeterBuilderFn, error) {
	return func(log logs.Logger) Meter {
		mm := &MultiMeter{}
		mm.wrapped = make([]Meter, 0, len(builders))
		for _, b := range builders {
			mm.wrapped = append(mm.wrapped, b(log))
		}
		return mm
	}, nil
}
