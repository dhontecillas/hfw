package config

import (
	"os"
	"strings"
)

func newMapConfFromEnv(prefix string, separator string) *MapConf {
	lowPrefix := strings.ToLower(prefix)
	lowSeparator := strings.ToLower(separator)

	if lowPrefix != "" {
		lowPrefix = lowPrefix + lowSeparator
	}

	envVars := os.Environ()
	mc := newMapConf(nil)
	for _, ev := range envVars {
		keyVal := strings.Split(ev, "=")
		if len(keyVal) != 2 {
			// TODO: this might be removes the ability to set
			// some value to an empty string ? put a test case for this
			continue
		}
		key := keyVal[0]
		val := keyVal[1]

		lkey := strings.ToLower(key)
		if !strings.HasPrefix(lkey, lowPrefix) {
			continue
		}
		path := strings.Split(lkey, lowSeparator)
		if len(path) <= 0 {
			// this should never happen
			// panic("WARNING: invariant not fullfilled")
			// TODO: emit a warning ?
			// TODO: check valid characters in paths ???
			// empty strings ??
			continue
		}
		nonEmptyPath := make([]string, 0, len(path)-1)
		// TODO: check if we want the prefix as the root key in the map
		for _, p := range path[1:] {
			if p == "" {
				continue
			}
			nonEmptyPath = append(nonEmptyPath, p)
		}
		if len(nonEmptyPath) == 0 {
			continue
		}
		mc.Set(nonEmptyPath, val)
	}
	return mc
}
