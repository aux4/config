package main

import (
	"encoding/json"
	"fmt"
)

// printValue prints a value in JSON format using standard json.Marshal with custom marshalers
func printValue(configValue any) {
	switch value := configValue.(type) {
	case Aux4Config:
		printValue(value.Config)
	case string:
		// In standalone context, print string without quotes
		fmt.Print(value)
	default:
		// Wrap value for deterministic JSON output and marshal
		wrapped := wrapForDeterministicJSON(value)
		data, err := json.Marshal(wrapped)
		if err != nil {
			fmt.Printf("Error marshaling JSON: %v", err)
			return
		}
		fmt.Print(string(data))
	}
}