package obs

import (
	"fmt"

	"github.com/dhontecillas/hfw/pkg/obs/logs"
	"github.com/dhontecillas/hfw/pkg/obs/metrics"
	"github.com/dhontecillas/hfw/pkg/obs/traces"
)

// Insighter is the top level object to report all
// metrics, traces and output logs.
type Insighter struct {
	L logs.Logger
	M metrics.Meter
	T traces.Tracer

	metricDefs metrics.MetricDefinitionList
}

// InsighterBuilderFn defines the function signature to create a new Insighter instance
type InsighterBuilderFn func() *Insighter

// NewInsighterBuilder constructs a new insighter
func NewInsighterBuilder(metricDefs metrics.MetricDefinitionList, lb logs.LoggerBuilderFn,
	mb metrics.MeterBuilderFn, tb traces.TracerBuilderFn) InsighterBuilderFn {

	return func() *Insighter {
		l := lb()
		mdefs, _ := metricDefs.CleanUp()
		m := mb(l)
		t := tb(l)
		return &Insighter{
			L:          l,
			M:          m,
			T:          t,
			metricDefs: mdefs,
		}
	}
}

// Clone creates a new insighter from the current one,
// clonning the logger but maintaining references to
// the metrics and traces instances.
func (i *Insighter) Clone() *Insighter {
	return &Insighter{
		L: i.L.Clone(),
		M: i.M,
		T: i.T,
	}
}

// Str sets a string label value for the underlying systems.
func (i *Insighter) Str(key, val string) *Insighter {
	i.L.Str(key, val)
	i.M.Str(key, val)
	i.T.Str(key, val)
	return i
}

// I64 sets an int label value for the underlying systems.
func (i *Insighter) I64(key string, val int64) *Insighter {
	i.L.I64(key, val)
	i.M.Str(key, fmt.Sprintf("%d", val))
	i.T.I64(key, val)
	return i
}

// F64 sets a float label value for the underlying systems.
func (i *Insighter) F64(key string, val float64) *Insighter {
	i.L.F64(key, val)
	i.M.Str(key, fmt.Sprintf("%f", val))
	i.T.F64(key, val)
	return i
}

// Bool sets a boolean label value for the underlying systems.
func (i *Insighter) Bool(key string, val bool) *Insighter {
	i.L.Bool(key, val)
	i.M.Str(key, fmt.Sprintf("%t", val))
	i.T.Bool(key, val)
	return i
}
