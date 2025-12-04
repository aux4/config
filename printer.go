package main

import (
	"encoding/json"
	"fmt"
)

// printValue prints a value in JSON format
func printValue(configValue interface{}) {
	printValueWithContext(configValue, false)
}

// printValueWithContext prints a value with context about whether it's inside an object
func printValueWithContext(configValue interface{}, inObject bool) {
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
	case []interface{}:
		// Handle arrays (including arrays of OrderedMap objects)
		fmt.Print("[")
		for i, item := range value {
			if i > 0 {
				fmt.Print(",")
			}
			printValueWithContext(item, true)
		}
		fmt.Print("]")
	case map[string]interface{}:
		// For regular maps, convert any nested OrderedMaps properly then marshal
		convertedMap := make(map[string]interface{})
		for key, val := range value {
			convertedMap[key] = convertOrderedMapToMap(val)
		}
		data, _ := json.Marshal(convertedMap)
		fmt.Print(string(data))
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