package config

import (
	"os"
	"strings"
)

type EnvLoader struct {
	mc MapConf
}

func newEnvConfLoader(prefix string, separator string) *EnvLoader {
	el := &EnvLoader{}

	lowPrefix := strings.ToLower(prefix)
	lowSeparator := strings.ToLower(separator)

	if lowPrefix != "" {
		lowPrefix = lowPrefix + lowSeparator
	}

	envVars := os.Environ()
	for _, ev := range envVars {
		lev := strings.ToLower(ev)
		if !strings.HasPrefix(lev, lowPrefix) {
			continue
		}
		keyVal := strings.Split(lev, "=")
		if len(keyVal) != 2 {
			// TODO: this might be removes the ability to set
			// some value to an empty string ? put a test case for this
			continue
		}
		key := keyVal[0]
		val := keyVal[1]
		path := strings.Split(key, lowSeparator)
		if len(path) <= 0 {
			// this should never happen
			panic("WARNING: invariant not fullfilled")
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
		el.mc.Set(nonEmptyPath, val)
	}

	return el
}
