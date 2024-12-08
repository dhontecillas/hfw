package attrs

import (
	"fmt"
)

type Attributable interface {
	// Str adds a tag to the message of type string.
	Str(key, val string)
	// I64 adds a tag to the message of type int64.
	I64(key string, val int64)
	// F64 adds a tag to the message of type float64.
	F64(key string, val float64)
	// Bool adds a tag to the message of type bool.
	Bool(key string, val bool)

	SetAttrs(vals map[string]interface{})
}

func SetAttr(a Attributable, k string, v interface{}) Attributable {
	switch vt := v.(type) {
	case string:
		a.Str(k, vt)
	case int64:
		a.I64(k, vt)
	case int:
		a.I64(k, int64(vt))
	case int8:
		a.I64(k, int64(vt))
	case int16:
		a.I64(k, int64(vt))
	case int32:
		a.I64(k, int64(vt))
	case uint:
		a.I64(k, int64(vt))
	case uint8:
		a.I64(k, int64(vt))
	case uint16:
		a.I64(k, int64(vt))
	case uint32:
		a.I64(k, int64(vt))
	case uint64:
		a.I64(k, int64(vt))
	case float64:
		a.F64(k, vt)
	case float32:
		a.F64(k, float64(vt))
	case bool:
		a.Bool(k, vt)
	default:
		a.Str(k, fmt.Sprintf("%+v", vt))
	}
	return a
}

func ApplyAttrs(a Attributable, attrMap map[string]interface{}) {
	if attrMap == nil {
		return
	}
	for k, v := range attrMap {
		SetAttr(a, k, v)
	}
}
