package main

import (
	"fmt"
	"os"
)

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
