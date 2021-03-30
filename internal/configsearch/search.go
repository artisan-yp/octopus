package configsearch

import (
	"github.com/k8s-practice/octopus/utils/cast"
)

// SearchPathInMap recursively searches for value for path in m map.
// Returns nil if not found.
func SearchPathInMap(m map[string]interface{}, path []string) interface{} {
	if len(path) == 0 || (len(path) == 1 && path[0] == "") {
		return m
	}

	if v, ok := m[path[0]]; ok {
		if len(path) == 1 {
			return v
		}

		switch v.(type) {
		// Nested case.
		case map[string]interface{}:
			return SearchPathInMap(v.(map[string]interface{}), path[1:])
		case map[interface{}]interface{}:
			return SearchPathInMap(cast.ToStringMap(v), path[1:])
		default:
			// Keywords that are not strings are not support.
			return nil
		}
	}

	return nil
}

// SetValueInMap sets value on the path in m map.
func SetValueInMap(m map[string]interface{}, path []string, value interface{}) {
	for _, key := range path[0 : len(path)-1] {
		if m2, ok := m[key]; ok {
			if m3, ok := m2.(map[string]interface{}); ok {
				m = m3
			}
		} else {
			m3 := make(map[string]interface{})
			m[key] = m3
			m = m3
		}
	}

	m[path[len(path)-1]] = value
}
