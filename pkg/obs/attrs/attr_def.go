package attrs

import (
	"errors"
	"fmt"
)

const (
	AttrTypeStr  string = "str"
	AttrTypeI64  string = "i64"
	AttrTypeF64  string = "f64"
	AttrTypeBool string = "bool"
)

var (
	// sentinel errors
	ErrAttrNotFound  = errors.New("ErrAttrNotFound")
	ErrAttrWrongType = errors.New("ErrAttrWrongType")
	ErrAttrBadName   = errors.New("ErrAttrWrongType")

	// AttrTypeNames maps a tag type number, with an string name
	attrValidTypes = map[string]bool{
		AttrTypeStr:  true,
		AttrTypeI64:  true,
		AttrTypeF64:  true,
		AttrTypeBool: true,
	}
)

type AttrDefinition struct {
	Name        string `json:"name"`
	StrAttrType string `json:"attr_type"`
}

type AttrDefinitionList []AttrDefinition

func (l AttrDefinitionList) CleanUp() (AttrDefinitionList, []error) {
	// uniqueNames contains the name, and at what index
	uniqueNames := map[string]int{}

	nl := make([]AttrDefinition, 0, len(l))
	var errList []error
	for idx, td := range l {
		if len(td.Name) == 0 {
			e := fmt.Errorf("attr #%d has empty name: %w", idx, ErrAttrBadName)
			errList = append(errList, e)
			continue
		}
		if ridx, ok := uniqueNames[td.Name]; ok {
			e := fmt.Errorf("attr #%d has same name as #%d: %w", idx, ridx, ErrAttrBadName)
			errList = append(errList, e)
			continue
		}
		// TODO: we can apply some "clean up" for valid character to the
		// attribute name.
		if len(td.StrAttrType) == 0 {
			td.StrAttrType = AttrTypeStr
		}
		if !attrValidTypes[td.StrAttrType] {
			e := fmt.Errorf("attr #%d invalid type '%s' : %w", idx,
				td.StrAttrType, ErrAttrWrongType)
			errList = append(errList, e)
			continue
		}
		uniqueNames[td.Name] = idx
		nl = append(nl, td)
	}

	return nl, errList
}

func (l AttrDefinitionList) Names() []string {
	names := make([]string, 0, len(l))
	for _, d := range l {
		names = append(names, d.Name)
	}
	return names
}
