package config

import (
	"fmt"
)

type MapConf struct {
	mi map[string]any
}

func newMapConf() *MapConf {
	return &MapConf{
		mi: make(map[string]any),
	}
}

func createOrOverwritePath(m map[string]any, path []string, val any) {
	lastElem := len(path) - 1
	for idx := 0; idx < lastElem; idx++ {
		el := path[idx]
		n := make(map[string]any)
		m[el] = n
		m = n
	}
	m[path[lastElem]] = val
}

// Set will create a path into the map to set a value.
// It will overwrite any other path that has a different type without
// emmiting an error.
func (m *MapConf) Set(path []string, val any) error {
	if m.mi == nil {
		m.mi = make(map[string]any)
	}

	cm := m.mi
	var ok bool
	var i any
	for idx, el := range path {
		i, ok = cm[el]
		if !ok {
			createOrOverwritePath(cm, path[idx:], val)
			return nil
		}
		cm, ok = i.(map[string]any)
		if !ok {
			// TODO: here we can emit a warning that we are overwriting
			// something that is not a map (that should be an option
			// when createing the MapConf)
			// return fmt.Errorf("not tree element %s at %d", el, idx)
			createOrOverwritePath(cm, path[idx:], val)
			return nil
		}
	}
	return nil
}

func (m *MapConf) Get(path []string) (any, error) {
	if m.mi == nil {
		return nil, fmt.Errorf("no values in map")
	}

	var cm map[string]any
	var ok bool
	var i any = m.mi
	for idx, el := range path {
		cm, ok = i.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("not map for %s at path idx %d", el, idx)
		}
		i, ok = cm[el]
		if !ok {
			return nil, fmt.Errorf("cannot find %s at path idx %d", el, idx)
		}
	}
	return i, nil
}
