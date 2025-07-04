package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type MapConf struct {
	mi map[string]any
}

var _ ConfLoader = (*MapConf)(nil)

var ErrEmptyMap = errors.New("empty map")

func newMapConf(data map[string]any) *MapConf {
	if data == nil {
		data = make(map[string]any)
	}
	// TODO: should we make a deep copy ?
	return &MapConf{
		mi: data,
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

func (m *MapConf) section(path []string) (*MapConf, error) {
	miAny, err := m.Get(path)
	if err != nil {
		return nil, fmt.Errorf("section %s not found: %s",
			strings.Join(path, "->"), err.Error())
	}
	mi, ok := miAny.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("section %s not map interface",
			strings.Join(path, "->"))
	}
	return &MapConf{
		mi: mi,
	}, nil
}

func (m *MapConf) Section(path []string) (ConfLoader, error) {
	return m.section(path)
}

// merges with override
func (m *MapConf) merge(a, b map[string]any) {
	for bkey, bval := range b {
		aval, ok := a[bkey]
		if !ok {
			// TODO: warning, we are not doing deep copy !
			aval = bval
			a[bkey] = aval
			continue
		}
		bMap, isBMap := bval.(map[string]any)
		if !isBMap {
			a[bkey] = bval
			continue
		}
		aMap, isAMap := aval.(map[string]any)
		if !isAMap {
			a[bkey] = bval
			continue
		}
		m.merge(aMap, bMap)
	}
}

func (m *MapConf) Merge(other *MapConf) {
	m.merge(m.mi, other.mi)
	fmt.Printf("\nmerged: \n%#v\n", m.mi)
}

func (m *MapConf) Parse(target any) error {
	if m.mi == nil {
		return ErrEmptyMap
	}
	b, err := json.Marshal(m.mi)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, target)
	if err != nil {
		return err
	}
	return nil
}
