package main

import (
	"bytes"
	"encoding/json"
	"strings"
)

// OrderedMap maintains the order of keys as they are added
type OrderedMap struct {
	Keys   []string                 `json:"Keys"`
	Values map[string]any `json:"Values"`
}

func newOrderedMap() *OrderedMap {
	return &OrderedMap{
		Keys:   make([]string, 0),
		Values: make(map[string]any),
	}
}

func (om *OrderedMap) Set(key string, value any) {
	if _, exists := om.Values[key]; !exists {
		om.Keys = append(om.Keys, key)
	}
	om.Values[key] = value
}

func (om *OrderedMap) Get(key string) (any, bool) {
	value, exists := om.Values[key]
	return value, exists
}

// MarshalJSON implements json.Marshaler interface for OrderedMap
// This ensures JSON output respects the key order defined in om.Keys
func (om *OrderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, key := range om.Keys {
		if i > 0 {
			buf.WriteByte(',')
		}

		// Marshal the key
		keyBytes, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')

		// Marshal the value - wrap in DeterministicMap if needed
		value := wrapForDeterministicJSON(om.Values[key])
		valueBytes, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// DeterministicMap wraps a regular map to ensure deterministic JSON marshaling
type DeterministicMap map[string]any

// MarshalJSON implements json.Marshaler interface for DeterministicMap
func (dm DeterministicMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	// Get keys and sort them for deterministic iteration
	var keys []string
	for key := range dm {
		keys = append(keys, key)
	}

	// Simple sort (bubble sort for simplicity)
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	// Marshal in sorted order
	for i, key := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}

		// Marshal the key
		keyBytes, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')

		// Marshal the value - wrap in DeterministicMap if needed
		value := wrapForDeterministicJSON(dm[key])
		valueBytes, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// wrapForDeterministicJSON wraps values to ensure deterministic JSON output
func wrapForDeterministicJSON(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return DeterministicMap(v)
	case []any:
		// Wrap array elements
		wrapped := make([]any, len(v))
		for i, item := range v {
			wrapped[i] = wrapForDeterministicJSON(item)
		}
		return wrapped
	default:
		return value
	}
}

// convertMapToOrderedMap converts a regular map to an OrderedMap recursively
func convertMapToOrderedMap(value any) any {
	switch v := value.(type) {
	case map[string]any:
		// For regular maps (from merge operations), we can't preserve original order
		// since Go maps have random iteration. Sort keys for deterministic order.
		om := newOrderedMap()

		// Get keys and sort them for deterministic iteration
		var keys []string
		for key := range v {
			keys = append(keys, key)
		}

		// Simple sort (bubble sort for simplicity)
		for i := 0; i < len(keys); i++ {
			for j := i + 1; j < len(keys); j++ {
				if keys[i] > keys[j] {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}

		// Set values in sorted key order
		for _, key := range keys {
			om.Set(key, convertMapToOrderedMap(v[key]))
		}
		return om
	case *OrderedMap:
		// OrderedMap is already in the right format, but recursively process values
		om := newOrderedMap()
		for _, key := range v.Keys {
			om.Set(key, convertMapToOrderedMap(v.Values[key]))
		}
		return om
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = convertMapToOrderedMap(item)
		}
		return result
	default:
		return v
	}
}

// convertOrderedMapToMap converts an OrderedMap back to a regular map recursively
func convertOrderedMapToMap(value any) any {
	switch v := value.(type) {
	case *OrderedMap:
		result := make(map[string]any)
		for _, key := range v.Keys {
			result[key] = convertOrderedMapToMap(v.Values[key])
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = convertOrderedMapToMap(item)
		}
		return result
	default:
		return v
	}
}

// getNestedValueFromOrderedMap retrieves a nested value from OrderedMap structures
func getNestedValueFromOrderedMap(value any, path string) (any, bool) {
	if path == "" {
		return value, true
	}

	keys := strings.Split(path, "/")
	current := value

	for _, key := range keys {
		if orderedMap, ok := current.(*OrderedMap); ok {
			var found bool
			current, found = orderedMap.Get(key)
			if !found {
				return nil, false
			}
		} else if regularMap, ok := current.(map[string]any); ok {
			var found bool
			current, found = regularMap[key]
			if !found {
				return nil, false
			}
		} else {
			return nil, false
		}
	}

	return current, true
}

// setNestedValueInOrderedMap sets a nested value in OrderedMap structures
func setNestedValueInOrderedMap(config any, path string, value any) any {
	// Convert to regular map, set the value, then convert back to OrderedMap
	mapConfig := convertOrderedMapToMap(config)

	// Use existing setNestedValue function
	setNestedValue(mapConfig.(map[string]any), path, value)

	// Convert back to OrderedMap
	return convertMapToOrderedMap(mapConfig)
}

// setNestedValue sets a nested value in a regular map
func setNestedValue(config map[string]any, path string, value any) {
	keys := strings.Split(path, "/")
	current := config

	for i, key := range keys {
		if i == len(keys)-1 {
			current[key] = value
		} else {
			if _, ok := current[key]; !ok {
				current[key] = make(map[string]any)
			}
			current = current[key].(map[string]any)
		}
	}
}