package main

import (
	"encoding/json"
	"io"
	"os"
)

func mergeConfig(root map[string]any) (map[string]any, error) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return map[string]any{}, err
	}

	var raw any

	var rawOrderedMap OrderedMap
	if err := json.Unmarshal(input, &rawOrderedMap); err != nil {
		var regularMap map[string]any
		if err := json.Unmarshal(input, &regularMap); err != nil {
			return root, err
		}
		raw = regularMap
	} else {
		raw = &rawOrderedMap
	}
	if om, ok := raw.(*OrderedMap); ok {
		if configValue, exists := om.Get("config"); exists {
			raw = configValue
		}
	} else if regularMap, ok := raw.(map[string]any); ok {
		if configValue, exists := regularMap["config"]; exists {
			raw = configValue
		}
	}

	source := root

	if _, ok := source["config"]; ok {
		source = source["config"].(map[string]any)
	}

	var merged map[string]any
	if om, ok := raw.(*OrderedMap); ok {
		merged = deepMergeWithOrderedMap(source, om)
	} else {
		incoming := convertToConfig(raw.(map[string]any))
		merged = deepMergeDeterministic(source, incoming)
	}

	return merged, nil
}

func deepMergeWithOrderedMap(sourceConfig map[string]any, incomingOrderedMap *OrderedMap) map[string]any {
	for _, key := range incomingOrderedMap.Keys {
		value := incomingOrderedMap.Values[key]

		if nestedOM, ok := value.(*OrderedMap); ok {
			if nestedSourceConfig, ok := sourceConfig[key].(map[string]any); ok {
				sourceConfig[key] = deepMergeWithOrderedMap(nestedSourceConfig, nestedOM)
			} else {
				newMap := make(map[string]any)
				sourceConfig[key] = deepMergeWithOrderedMap(newMap, nestedOM)
			}
		} else if nestedIncomingConfig, ok := value.(map[string]any); ok {
			if nestedSourceConfig, ok := sourceConfig[key].(map[string]any); ok {
				sourceConfig[key] = deepMergeDeterministic(nestedSourceConfig, nestedIncomingConfig)
			} else {
				sourceConfig[key] = nestedIncomingConfig
			}
		} else {
			sourceConfig[key] = value
		}
	}

	return sourceConfig
}

func deepMergeDeterministic(sourceConfig map[string]any, incomingConfig map[string]any) map[string]any {
	var keys []string
	for key := range incomingConfig {
		keys = append(keys, key)
	}

	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	for _, key := range keys {
		value := incomingConfig[key]
		if nestedIncomingConfig, ok := value.(map[string]any); ok {
			if nestedSourceConfig, ok := sourceConfig[key].(map[string]any); ok {
				sourceConfig[key] = deepMergeDeterministic(nestedSourceConfig, nestedIncomingConfig)
			} else {
				sourceConfig[key] = nestedIncomingConfig
			}
		} else {
			sourceConfig[key] = value
		}
	}

	return sourceConfig
}
