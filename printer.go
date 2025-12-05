package main

import (
	"encoding/json"
	"fmt"
)

// printValue prints a value in JSON format
func printValue(configValue any) {
	printValueWithContext(configValue, false)
}

// printValueWithContext prints a value with context about whether it's inside an object
func printValueWithContext(configValue any, inObject bool) {
	switch value := configValue.(type) {
	case *OrderedMap:
		fmt.Print("{")
		for i, key := range value.Keys {
			if i > 0 {
				fmt.Print(",")
			}
			keyJSON, _ := json.Marshal(key)
			fmt.Printf("%s:", keyJSON)
			printValueWithContext(value.Values[key], true)
		}
		fmt.Print("}")
	case []any:
		// Handle arrays (including arrays of OrderedMap objects)
		fmt.Print("[")
		for i, item := range value {
			if i > 0 {
				fmt.Print(",")
			}
			printValueWithContext(item, true)
		}
		fmt.Print("]")
	case map[string]any:
		// For regular maps, print in sorted order for deterministic output
		fmt.Print("{")

		// Get keys and sort them for deterministic iteration
		var keys []string
		for key := range value {
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

		// Print in sorted order
		for i, key := range keys {
			if i > 0 {
				fmt.Print(",")
			}
			keyJSON, _ := json.Marshal(key)
			fmt.Printf("%s:", keyJSON)
			printValueWithContext(value[key], true)
		}
		fmt.Print("}")
	case Aux4Config:
		printValueWithContext(value.Config, inObject)
	case string:
		if inObject {
			// In object context, strings need to be properly quoted for valid JSON
			data, _ := json.Marshal(value)
			fmt.Print(string(data))
		} else {
			// In standalone context, print string without quotes
			fmt.Print(value)
		}
	default:
		data, _ := json.Marshal(value)
		fmt.Print(string(data))
	}
}