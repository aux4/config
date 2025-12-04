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

	var value interface{}
	var found bool

	if path == "" {
		value = aux4Config.Config
	} else {
		value, found = getNestedValueFromOrderedMap(aux4Config.Config, path)
		if !found {
			value = make(map[string]interface{})
		}
	}

	// Convert value to map for merging
	var mapValue map[string]interface{}
	switch v := value.(type) {
	case *OrderedMap:
		mapValue = v.Values
	case map[string]interface{}:
		mapValue = v
	default:
		mapValue = make(map[string]interface{})
	}

	mergedConfig, err := mergeConfig(mapValue)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error merging config: %v\n", err)
		os.Exit(1)
	}

	if path == "" {
		aux4Config.Config = mergedConfig
	} else {
		aux4Config.Config = setNestedValueInOrderedMap(aux4Config.Config, path, mergedConfig)
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
func mergeConfig(root map[string]interface{}) (map[string]interface{}, error) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return map[string]interface{}{}, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(input, &raw); err != nil {
		return root, err
	}

	if _, ok := raw["config"]; ok {
		raw = raw["config"].(map[string]interface{})
	}

	source := root

	if _, ok := source["config"]; ok {
		source = source["config"].(map[string]interface{})
	}

	incoming := convertToConfig(raw)

	merged := deepMerge(source, incoming)

	return merged, nil
}

// deepMerge performs a deep merge of two maps
func deepMerge(sourceConfig map[string]interface{}, incomingConfig map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy all values from source
	for key, value := range sourceConfig {
		result[key] = value
	}

	// Merge incoming values
	for key, incomingValue := range incomingConfig {
		if existingValue, exists := result[key]; exists {
			if existingMap, ok := existingValue.(map[string]interface{}); ok {
				if incomingMap, ok := incomingValue.(map[string]interface{}); ok {
					result[key] = deepMerge(existingMap, incomingMap)
					continue
				}
			}
		}
		result[key] = incomingValue
	}

	return result
}