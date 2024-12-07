package obs

import (
	"fmt"

	"github.com/dhontecillas/hfw/pkg/obs/attrs"
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
}

// InsighterBuilderFn defines the function signature to create a new Insighter instance
type InsighterBuilderFn func() *Insighter

// NewInsighterBuilder constructs a new insighter
func NewInsighterBuilder(lb logs.LoggerBuilderFn,
	mb metrics.MeterBuilderFn, tb traces.TracerBuilderFn) InsighterBuilderFn {

	return func() *Insighter {
		l := lb()
		m := mb(l) // we might end up with to many logs if it has errors ?
		t := tb(l)
		return &Insighter{
			L: l,
			M: m,
			T: t,
		}
	}
}

// Clone creates a new insighter from the current one,
// clonning the logger but maintaining references to
// the metrics and traces instances.
func (i *Insighter) Clone() *Insighter {
	return &Insighter{
		L: i.L.Clone(),
		M: i.M.Clone(),
		T: i.T, // tracer should not be cloned, a new span might be created
	}
}

// Str sets a string label value for the underlying systems.
func (i *Insighter) Str(key, val string) {
	i.L.Str(key, val)
	i.M.Str(key, val)
	i.T.Str(key, val)
}

// I64 sets an int label value for the underlying systems.
func (i *Insighter) I64(key string, val int64) {
	i.L.I64(key, val)
	i.M.Str(key, fmt.Sprintf("%d", val))
	i.T.I64(key, val)
}

// F64 sets a float label value for the underlying systems.
func (i *Insighter) F64(key string, val float64) {
	i.L.F64(key, val)
	i.M.Str(key, fmt.Sprintf("%f", val))
	i.T.F64(key, val)
}

// Bool sets a boolean label value for the underlying systems.
func (i *Insighter) Bool(key string, val bool) {
	i.L.Bool(key, val)
	i.M.Str(key, fmt.Sprintf("%t", val))
	i.T.Bool(key, val)
}

func (i *Insighter) SetAttrs(attrMap map[string]interface{}) {
	attrs.ApplyAttrs(i, attrMap)
}
