package traces

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
)

func otelAttrs(attrMap map[string]interface{}) []attribute.KeyValue {
	res := make([]attribute.KeyValue, 0, len(attrMap))

	for k, v := range attrMap {
		switch vt := v.(type) {
		case string:
			res = append(res, attribute.String(k, vt))
		case int64:
			res = append(res, attribute.Int64(k, vt))
		case int:
			res = append(res, attribute.Int64(k, int64(vt)))
		case int8:
			res = append(res, attribute.Int64(k, int64(vt)))
		case int16:
			res = append(res, attribute.Int64(k, int64(vt)))
		case int32:
			res = append(res, attribute.Int64(k, int64(vt)))
		case uint:
			res = append(res, attribute.Int64(k, int64(vt)))
		case uint8:
			res = append(res, attribute.Int64(k, int64(vt)))
		case uint16:
			res = append(res, attribute.Int64(k, int64(vt)))
		case uint32:
			res = append(res, attribute.Int64(k, int64(vt)))
		case uint64:
			res = append(res, attribute.Int64(k, int64(vt)))
		case float64:
			res = append(res, attribute.Float64(k, float64(vt)))
		case float32:
			res = append(res, attribute.Float64(k, float64(vt)))
		case bool:
			res = append(res, attribute.Bool(k, vt))
		case []string:
			res = append(res, attribute.StringSlice(k, vt))
		default:
			res = append(res, attribute.String(k, fmt.Sprintf("%+v", vt)))
		}
	}
	return res
}
