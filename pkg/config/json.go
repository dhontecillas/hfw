package config

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func newMapConfFromJSON(content []byte) (*MapConf, error) {
	var m map[string]any
	var a []any
	mErr := json.Unmarshal(content, &m)
	aErr := json.Unmarshal(content, &a)

	if mErr != nil && aErr != nil {
		return nil, fmt.Errorf("cannot decode content")
	}

	// if it is an array, we convert it to the map[string]interface
	// using strings to later on only have a way to iterate the keys
	if mErr != nil {
		m = make(map[string]any, len(a))
		for idx := range a {
			m[strconv.Itoa(idx)] = a[idx]
		}
	}

	return newMapConf(m), nil
}
