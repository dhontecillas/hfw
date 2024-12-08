package obs

import (
	"context"

	"github.com/dhontecillas/hfw/pkg/obs/logs"
	"github.com/dhontecillas/hfw/pkg/obs/metrics"
	"github.com/dhontecillas/hfw/pkg/obs/traces"
)

type insighterContextKey string

const (
	// InsighterContextKey is the key used to store the
	// insighter in a context for a given request.
	InsighterContextKey insighterContextKey = "Insighter"
)

func nopInsighter() *Insighter {
	nopLoggerBuilder := logs.NewNopLoggerBuilder()
	l := nopLoggerBuilder()

	nopMeterBuilder, err := metrics.NewNopMeterBuilder()
	if err != nil {
		l.Err(err, "cannot create nop builder", nil)
	}
	nopTracerBuilder := traces.NewNopTracerBuilder()
	nopInsighterBuilder := NewInsighterBuilder(
		nopLoggerBuilder, nopMeterBuilder, nopTracerBuilder)

	return nopInsighterBuilder()
}

// InsighterFromContext retrieves an insighter from the current
// context in case there is one attached to it.
func InsighterFromContext(ctx context.Context) *Insighter {
	v := ctx.Value(InsighterContextKey)
	if v == nil {
		return nopInsighter()
	}
	return v.(*Insighter)
}

// InsighterWithContext attaches an Insighter to a Context.
func InsighterWithContext(ctx context.Context, ins *Insighter) context.Context {
	if ins == nil {
		return ctx
	}
	return context.WithValue(ctx, InsighterContextKey, ins)
}
