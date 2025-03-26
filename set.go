// Copyright 2025 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"maps"
	"slices"
)

// stringsSet represents the classic "set" data structure, and contains strings.
type stringsSet map[string]bool

func (s stringsSet) values() []string {
	return slices.Sorted(maps.Keys(s))
}

func (s stringsSet) add(value string) {
	s[value] = true
}

func (s stringsSet) contains(value string) bool {
	_, exists := s[value]
	return exists
}

func (s stringsSet) union(other stringsSet) stringsSet {
	result := maps.Clone(s)
	maps.Copy(result, other)
	return result
}

func (s stringsSet) difference(other stringsSet) stringsSet {
	result := maps.Clone(s)
	maps.DeleteFunc(result, func(value string, _ bool) bool {
		return other.contains(value)
	})
	return result
}
