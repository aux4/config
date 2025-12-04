package main

import "strings"

// OrderedMap maintains the order of keys as they are added
type OrderedMap struct {
	Keys   []string                 `json:"Keys"`
	Values map[string]interface{} `json:"Values"`
}

func newOrderedMap() *OrderedMap {
	return &OrderedMap{
		Keys:   make([]string, 0),
		Values: make(map[string]interface{}),
	}
}

func (om *OrderedMap) Set(key string, value interface{}) {
	if _, exists := om.Values[key]; !exists {
		om.Keys = append(om.Keys, key)
	}
	om.Values[key] = value
}

func (om *OrderedMap) Get(key string) (interface{}, bool) {
	value, exists := om.Values[key]
	return value, exists
}

// convertMapToOrderedMap converts a regular map to an OrderedMap recursively
func convertMapToOrderedMap(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		// For regular maps (from merge operations), we can't preserve original order
		// since Go maps have random iteration. We'll just use the iteration order.
		om := newOrderedMap()
		for key, val := range v {
			om.Set(key, convertMapToOrderedMap(val))
		}
		return om
	case *OrderedMap:
		// OrderedMap is already in the right format, but recursively process values
		om := newOrderedMap()
		for _, key := range v.Keys {
			om.Set(key, convertMapToOrderedMap(v.Values[key]))
		}
		return om
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = convertMapToOrderedMap(item)
		}
		return result
	default:
		return v
	}
}

// convertOrderedMapToMap converts an OrderedMap back to a regular map recursively
func convertOrderedMapToMap(value interface{}) interface{} {
	switch v := value.(type) {
	case *OrderedMap:
		result := make(map[string]interface{})
		for _, key := range v.Keys {
			result[key] = convertOrderedMapToMap(v.Values[key])
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = convertOrderedMapToMap(item)
		}
		return result
	default:
		return v
	}
}

// getNestedValueFromOrderedMap retrieves a nested value from OrderedMap structures
func getNestedValueFromOrderedMap(value interface{}, path string) (interface{}, bool) {
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
		} else if regularMap, ok := current.(map[string]interface{}); ok {
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
func setNestedValueInOrderedMap(config interface{}, path string, value any) interface{} {
	// Convert to regular map, set the value, then convert back to OrderedMap
	mapConfig := convertOrderedMapToMap(config)

	// Use existing setNestedValue function
	setNestedValue(mapConfig.(map[string]interface{}), path, value)

	// Convert back to OrderedMap
	return convertMapToOrderedMap(mapConfig)
}

// setNestedValue sets a nested value in a regular map
func setNestedValue(config map[string]interface{}, path string, value any) {
	keys := strings.Split(path, "/")
	current := config

	for i, key := range keys {
		if i == len(keys)-1 {
			current[key] = value
		} else {
			if _, ok := current[key]; !ok {
				current[key] = make(map[string]interface{})
			}
			current = current[key].(map[string]interface{})
		}
	}
}