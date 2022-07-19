package metrics

import (
	"fmt"

	"github.com/dhontecillas/hfw/pkg/obs/logs"
)

// Type definitions for the different kind of metrics available.
const (
	MetricTypeMonotonicCounter = iota
	MetricTypeUpDownCounter
	MetricTypeHistogram
	MetricTypeDistribution

	MetricTypeInvalid
)

// we allow to special types of metrics for specialized
// implementations ? :
const (
	// MetricTypeExtension is the minimum value for a custom metric type
	MetricTypeExtension = iota + 100
)

// Def contains the information for a metric definition.
// It contains the name of the metric, the type, and the labels
// that can be applied to this metric.
type Def struct {
	Name       string
	MetricType int
	Labels     []string
}

func (md Def) copy() Def {
	lbls := make([]string, len(md.Labels))
	copy(lbls, md.Labels)
	return Def{
		Name:       md.Name,
		MetricType: md.MetricType,
		Labels:     lbls,
	}
}

// Defs defines a list of metric definitions.
type Defs []Def

// Catalog contains the list of defined metrics.
type Catalog struct {
	defs  Defs
	index map[string]int
}

// Clean resets the MetricsCatalog leaving it empty.
func (md Defs) Clean(l logs.Logger) Catalog {
	if len(md) == 0 {
		return Catalog{
			index: map[string]int{},
		}
	}

	imd := Catalog{
		defs:  make(Defs, 0, len(md)),
		index: make(map[string]int, len(md)),
	}

	for idx, def := range md {
		if len(def.Name) == 0 {
			l.Warn(fmt.Sprintf("found empty name at idx %d", idx))
			continue
		}
		didx, ok := imd.index[def.Name]
		if ok {
			l.Warn(fmt.Sprintf("found duplicate %s (cur: %d, new: %d) at idx %d",
				def.Name, imd.defs[didx].MetricType, def.MetricType, idx))
			continue
		}

		if def.MetricType < MetricTypeMonotonicCounter || def.MetricType >= MetricTypeInvalid {
			l.Warn(fmt.Sprintf("found non 'standard' type %d for %s at idx %d",
				def.MetricType, def.Name, idx))
		}

		uniqueLabels := map[string]bool{}
		for _, l := range def.Labels {
			uniqueLabels[l] = true
		}
		labelsCopy := make([]string, 0, len(uniqueLabels))
		for k := range uniqueLabels {
			labelsCopy = append(labelsCopy, k)
		}

		imd.index[def.Name] = len(imd.defs)
		imd.defs = append(imd.defs, Def{
			Name:       def.Name,
			MetricType: def.MetricType,
			Labels:     labelsCopy,
		})
	}
	return imd
}

// Merge adds metrics definitions from another catalog, with
// option of overriding an existing one if the name of the matric
// already exists.
func (md Defs) Merge(other Defs, override bool) Defs {
	byName := make(map[string]int, len(md))
	newMD := make(Defs, 0, len(md)+len(other))
	for idx, m := range md {
		byName[m.Name] = idx
		newMD = append(newMD, m.copy())
	}
	for _, om := range other {
		idx, ok := byName[om.Name]
		if ok {
			if override {
				newMD[idx] = om.copy()
			}
		} else {
			newMD = append(newMD, om.copy())
		}
	}
	return newMD
}

// Def returns a metric definition by its name.
func (mc *Catalog) Def(name string) (*Def, int) {
	idx, ok := mc.index[name]
	if !ok || idx < 0 {
		return nil, -1
	}
	return &mc.defs[idx], idx
}
