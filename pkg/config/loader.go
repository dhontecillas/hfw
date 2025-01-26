package config

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type ConfLoader interface {
	Section(name string) (ConfLoader, error)
	Load(name string, target *any) error
}

type JSONConfLoader struct {
	mapContent   map[string]any
	arrayContent []any
}

func newJSONConfLoader(content []byte) (*JSONConfLoader, error) {
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

	return &JSONConfLoader{
		mapContent:   m,
		arrayContent: a,
	}, nil
}

func (l *JSONConfLoader) Section(name string) (ConfLoader, error) {
	return nil, fmt.Errorf("not implemented")
}

func (l *JSONConfLoader) Load(name string, target *any) error {
	return fmt.Errorf("not implemented")
}
