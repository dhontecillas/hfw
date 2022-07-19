package obs

import (
	"errors"
	"fmt"

	"github.com/dhontecillas/hfw/pkg/obs/logs"
	"github.com/dhontecillas/hfw/pkg/obs/metrics"
	"github.com/dhontecillas/hfw/pkg/obs/traces"
)

// TODO: unify naming, remove all `Tag` and use only
// labels, because it might be misleading.

// TagType defines the type to be used for a label
const (
	TagTypeNone = iota
	TagTypeStr
	TagTypeI64
	TagTypeF64
	TagTypeBool
	TagTypeInvalid
)

var (
	// sentinel errors
	errNotFound  = errors.New("ErrInsightsTagNotFound")
	errWrongType = errors.New("ErrInsightsWrongType")

	// TagTypeNames maps a tag type number, with an string name
	TagTypeNames = map[int]string{
		TagTypeStr:  "TagTypeStr",
		TagTypeI64:  "TagTypeI64",
		TagTypeF64:  "TagTypeF64",
		TagTypeBool: "TagTypeBool",
	}
)

// Insighter is the top level object to report all
// metrics, traces and output logs.
type Insighter struct {
	L logs.Logger
	M metrics.Meter
	T traces.Tracer

	tagDefs []TagDefinition
}

// InsighterBuilderFn defines the function signature to create a new Insighter instance
type InsighterBuilderFn func() *Insighter

// TagDefinition defines the type for a tagname, and to
// which subsystems we should set the tag. It allows to
// have all subsytems to false, so we allow a value to
// be set, even when it is not going to be sent anywhere
// per system configuration.
// For example, in production mode we cannot set some value
// only used in beta, or dev environments.
type TagDefinition struct {
	Name    string
	TagType int
	ToL     bool
	ToM     bool
	ToT     bool
}

// copyAndCleanupTagTargets checks that tag types are valid, and removes
// duplicates from the tag definition. The returned value is a new slice
// that can be saved inside the Insighter.
func copyAndCleanupTagTargets(tagDefs []TagDefinition) []TagDefinition {
	uniqueNames := map[string]bool{}
	tDefsCopy := make([]TagDefinition, 0, len(tagDefs))
	for _, td := range tagDefs {
		if len(td.Name) == 0 || td.TagType <= TagTypeNone || td.TagType >= TagTypeInvalid {
			continue
		}
		if _, ok := uniqueNames[td.Name]; ok {
			continue // or panic because of bad config ?
		}
		tDefsCopy = append(tDefsCopy, td)
	}
	return tDefsCopy
}

// NewInsighterBuilder constructs a new insighter
func NewInsighterBuilder(tagDefs []TagDefinition, lb logs.LoggerBuilderFn,
	mb metrics.MeterBuilderFn, tb traces.TracerBuilderFn) InsighterBuilderFn {
	tDefs := copyAndCleanupTagTargets(tagDefs)
	return func() *Insighter {
		l := lb()
		m := mb(l)
		t := tb(l)
		return &Insighter{
			L:       l,
			M:       m,
			T:       t,
			tagDefs: tDefs,
		}
	}
}

// Clone creates a new insighter from the current one,
// clonning the logger but maintaining references to
// the metrics and traces instances.
func (i *Insighter) Clone() *Insighter {
	return &Insighter{
		L:       i.L.Clone(),
		M:       i.M,
		T:       i.T,
		tagDefs: i.tagDefs,
	}
}

// CloneWith creates a new insighter keeping the same tag definitions
// but replacing the logger, metter and tracer if provided. If any of
// those fields is null, it will clone the existing one.
func (i *Insighter) CloneWith(l logs.Logger, m metrics.Meter, t traces.Tracer) *Insighter {
	if l == nil {
		l = i.L.Clone()
	}
	if m == nil {
		m = i.M
	}
	if t == nil {
		t = i.T
	}
	return &Insighter{
		L:       l,
		M:       m,
		T:       t,
		tagDefs: i.tagDefs,
	}
}

func (i *Insighter) findTagDef(name string, tagType int) (*TagDefinition, error) {
	for _, t := range i.tagDefs {
		if t.Name == name {
			if t.TagType != tagType {
				return nil, errWrongType
			}
			return &t, nil
		}
	}
	return nil, errNotFound
}

func (i *Insighter) logTagError(err error, key string, tagType int) *Insighter {
	i.L.ErrMsg(err, "cannot set insights tag").Str("tag", key).Str(
		"tagtype", TagTypeNames[tagType]).Send()
	return i
}

// Str sets a string label value for the underlying systems.
func (i *Insighter) Str(key, val string) *Insighter {
	t, err := i.findTagDef(key, TagTypeStr)
	if err != nil {
		return i.logTagError(err, key, TagTypeStr)
	}
	if t.ToL {
		i.L.Str(key, val)
	}
	if t.ToM {
		i.M.Str(key, val)
	}
	if t.ToT {
		i.T.Str(key, val)
	}
	return i
}

// I64 sets an int label value for the underlying systems.
func (i *Insighter) I64(key string, val int64) *Insighter {
	t, err := i.findTagDef(key, TagTypeI64)
	if err != nil {
		return i.logTagError(err, key, TagTypeI64)
	}
	if t.ToL {
		i.L.I64(key, val)
	}
	if t.ToM {
		i.M.Str(key, fmt.Sprintf("%d", val))
	}
	if t.ToT {
		i.T.I64(key, val)
	}
	return i
}

// F64 sets a float label value for the underlying systems.
func (i *Insighter) F64(key string, val float64) *Insighter {
	t, err := i.findTagDef(key, TagTypeF64)
	if err != nil {
		return i.logTagError(err, key, TagTypeF64)
	}
	if t.ToL {
		i.L.F64(key, val)
	}
	if t.ToM {
		i.M.Str(key, fmt.Sprintf("%f", val))
	}
	if t.ToT {
		i.T.F64(key, val)
	}
	return i
}

// Bool sets a boolean label value for the underlying systems.
func (i *Insighter) Bool(key string, val bool) *Insighter {
	t, err := i.findTagDef(key, TagTypeBool)
	if err != nil {
		return i.logTagError(err, key, TagTypeBool)
	}
	if t.ToL {
		i.L.Bool(key, val)
	}
	if t.ToM {
		i.M.Str(key, fmt.Sprintf("%t", val))
	}
	if t.ToT {
		i.T.Bool(key, val)
	}
	return i
}
