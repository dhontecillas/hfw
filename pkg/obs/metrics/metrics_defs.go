package metrics

import (
	"errors"
	"fmt"

	"github.com/dhontecillas/hfw/pkg/obs/attrs"
)

// Type definitions for the different kind of metrics available.
const (
	MetricTypeMonotonicCounter string = "counter"
	MetricTypeHistogram        string = "histogram"
	MetricTypeDistribution     string = "distribution"
	MetricTypeUpDownCounter    string = "updowncounter"

	MetricTypeInvalid
)

// we allow to special types of metrics for specialized
// implementations ? :
const (
	// MetricTypeExtension is the minimum value for a custom metric type
	MetricTypeExtension = iota + 100
)

var (
	ErrMetricNotFound  = errors.New("ErrMetricNotFound")
	ErrMetricWrongType = errors.New("ErrMetricWrongType")
	ErrMetricBadName   = errors.New("ErrMetricWrongType")

	validMetricTypes = map[string]bool{
		MetricTypeMonotonicCounter: true,
		MetricTypeHistogram:        true,
		MetricTypeDistribution:     true,
		MetricTypeUpDownCounter:    true,
	}
)

// Def contains the information for a metric definition.
// It contains the name of the metric, the type, and the labels
// that can be applied to this metric.
type MetricDefinition struct {
	Name         string                   `json:"name"`
	Units        string                   `json:"units"`
	MetricType   string                   `json:"metric_type"`
	Attributes   attrs.AttrDefinitionList `json:"attributes"`
	DefaultValue *string                  `json:"default"` // in case we want to set a fixed value if not set
	Extra        map[string]interface{}   `json:"extra"`   // in extra we can define the histogram
}

func (d *MetricDefinition) CleanUp() (*MetricDefinition, []error) {
	if len(d.Name) == 0 {
		e := fmt.Errorf("empty metric name: %w", ErrMetricBadName)
		return nil, []error{e}
	}

	if validMetricTypes[d.MetricType] {
		e := fmt.Errorf("unknown metric type: %w", ErrMetricWrongType)
		return nil, []error{e}
	}

	// we do not apply sanitization to the metric
	attrs, errs := d.Attributes.CleanUp()

	// TODO: the default value should be sanitized to make sure is a valid
	// value

	nd := &MetricDefinition{
		Name:         d.Name,
		Units:        d.Units,
		MetricType:   d.MetricType,
		Attributes:   attrs,
		DefaultValue: d.DefaultValue,
		Extra:        d.Extra,
	}
	return nd, errs
}

func (d *MetricDefinition) Copy(other *MetricDefinition) {
	if other == nil {
		return
	}
	*d = *other
}

type MetricDefinitionList []*MetricDefinition

func (l MetricDefinitionList) CleanUp() (MetricDefinitionList, []error) {
	uniqueNames := map[string]int{}
	errList := make([]error, 0, 16)
	nl := make(MetricDefinitionList, 0, len(l))
	for idx, md := range l {
		// TOOD: we can apply some name sanitization for the metric
		if ridx, ok := uniqueNames[md.Name]; ok {
			e := fmt.Errorf("metric #%d has same name as #%d: %w", idx, ridx, ErrMetricBadName)
			errList = append(errList, e)
			continue
		}

		nmd, errs := md.CleanUp()
		// we wrap all the errors with its own index
		for _, e := range errs {
			ne := fmt.Errorf("metric #%d: %w", idx, e)
			errList = append(errList, ne)
		}
		if nmd == nil {
			continue
		}

		uniqueNames[md.Name] = idx
		nl = append(nl, nmd)
	}

	return nl, errList
}

// Merge adds metrics definitions from another catalog, with
// option of overriding an existing one if the name of the matric
// already exists.
func (l MetricDefinitionList) Merge(other MetricDefinitionList, override bool) MetricDefinitionList {
	cleanOther, _ := other.CleanUp()

	toAppend := make(MetricDefinitionList, 0, len(other))
	// n 2  search , we assume small amount of entries:
	for _, o := range cleanOther {
		if fo, _, err := l.Def(o.Name); err != nil {
			if override {
				*fo = *o
			}
			continue
		}
		toAppend = append(toAppend, o)
	}
	merged := l
	return append(merged, toAppend...)
}

// Def returns a metric definition by its name.
func (l MetricDefinitionList) Def(name string, validTypes ...string) (*MetricDefinition, int, error) {
	foundIdx := -1
	for idx, md := range l {
		if md.Name == name {
			foundIdx = idx
			break
		}
	}

	if foundIdx < 0 {
		return nil, -1, ErrMetricNotFound
	}

	found := l[foundIdx]
	if len(validTypes) == 0 {
		return found, foundIdx, nil
	}

	for _, vt := range validTypes {
		if found.MetricType == vt {
			return found, foundIdx, nil
		}
	}
	return nil, -1, ErrMetricWrongType
}
