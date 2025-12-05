package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// getConfig handles the get command
func getConfig(aux4Config Aux4Config) {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Missing parameters\n")
		os.Exit(1)
	}
	path := os.Args[3]
	value, found := getNestedValueFromOrderedMap(aux4Config.Config, path)
	if found {
		printValue(value)
	}
}

// setConfig handles the set command
func setConfig(aux4Config Aux4Config, configFile string) {
	if len(os.Args) < 5 {
		fmt.Fprintf(os.Stderr, "Usage: aux4 config set <path> <value>\n")
		os.Exit(1)
	}

	path, value := os.Args[3], os.Args[4]
	aux4Config.Config = setNestedValueInOrderedMap(aux4Config.Config, path, value)

	if err := saveConfig(configFile, aux4Config); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}
}

// mergeAux4Config handles the merge command
func mergeAux4Config(aux4Config Aux4Config, configFile string) {
	if len(os.Args) < 5 {
		fmt.Fprintf(os.Stderr, "Missing parameters\n")
		os.Exit(1)
	}
	path := os.Args[3]
	save := os.Args[4] == "true"


	var value any
	var found bool

	if path == "" {
		value = aux4Config.Config
	} else {
		value, found = getNestedValueFromOrderedMap(aux4Config.Config, path)
		if !found {
			value = make(map[string]any)
		}
	}

	// Convert value to map for merging
	var mapValue map[string]any
	switch v := value.(type) {
	case *OrderedMap:
		mapValue = v.Values
	case map[string]any:
		mapValue = v
	default:
		mapValue = make(map[string]any)
	}

	mergedConfig, err := mergeConfig(mapValue)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error merging config: %v\n", err)
		os.Exit(1)
	}

	// Convert merged config back to OrderedMap to preserve order
	var mergedOrderedMap any
	if path == "" {
		mergedOrderedMap = convertMapToOrderedMap(mergedConfig)
		aux4Config.Config = mergedOrderedMap
	} else {
		mergedOrderedMap = convertMapToOrderedMap(mergedConfig)
		aux4Config.Config = setNestedValueInOrderedMap(aux4Config.Config, path, mergedOrderedMap)
	}

	if save {
		if err := saveConfig(configFile, aux4Config); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving merged config: %v\n", err)
			os.Exit(1)
		}
	} else {
		printValue(aux4Config)
	}
}

// mergeConfig reads from stdin and merges with existing config
func mergeConfig(root map[string]any) (map[string]any, error) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return map[string]any{}, err
	}

	var raw any

	// Parse JSON preserving order (stdin might not have "config" wrapper)
	rawOrderedMap, err := parseJSONValue(string(input))
	if err != nil {
		// If JSON parsing fails, fall back to regular JSON parsing
		var regularMap map[string]any
		if err := json.Unmarshal(input, &regularMap); err != nil {
			return root, err
		}
		raw = regularMap
	} else {
		// Keep as OrderedMap to preserve order during merge
		raw = rawOrderedMap
	}

	// Handle extracting config from either OrderedMap or regular map
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

	// Merge preserving order
	var merged map[string]any
	if om, ok := raw.(*OrderedMap); ok {
		merged = deepMergeWithOrderedMap(source, om)
	} else {
		incoming := convertToConfig(raw.(map[string]any))
		merged = deepMergeDeterministic(source, incoming)
	}

	return merged, nil
}


// deepMergeWithOrderedMap merges an OrderedMap into a regular map, preserving the OrderedMap's key order
func deepMergeWithOrderedMap(sourceConfig map[string]any, incomingOrderedMap *OrderedMap) map[string]any {
	// Iterate using the OrderedMap's original key order
	for _, key := range incomingOrderedMap.Keys {
		value := incomingOrderedMap.Values[key]

		// If value is also an OrderedMap, handle recursively
		if nestedOM, ok := value.(*OrderedMap); ok {
			if nestedSourceConfig, ok := sourceConfig[key].(map[string]any); ok {
				sourceConfig[key] = deepMergeWithOrderedMap(nestedSourceConfig, nestedOM)
			} else {
				// Convert OrderedMap to regular map preserving order
				newMap := make(map[string]any)
				sourceConfig[key] = deepMergeWithOrderedMap(newMap, nestedOM)
			}
		} else if nestedIncomingConfig, ok := value.(map[string]any); ok {
			if nestedSourceConfig, ok := sourceConfig[key].(map[string]any); ok {
				// Use OrderedMap merge to preserve any nested OrderedMaps
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

// deepMergeDeterministic performs a deep merge with deterministic key iteration
func deepMergeDeterministic(sourceConfig map[string]any, incomingConfig map[string]any) map[string]any {
	// Get keys and sort them for deterministic iteration
	var keys []string
	for key := range incomingConfig {
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

	// Iterate in sorted order
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