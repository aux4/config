package main

import (
	"encoding/json"
	"fmt"
)

func printValue(configValue any) {
	switch value := configValue.(type) {
	case Aux4Config:
		printValue(value.Config)
	case string:
		fmt.Print(value)
	default:
		wrapped := wrapForDeterministicJSON(value)
		data, err := json.Marshal(wrapped)
		if err != nil {
			fmt.Printf("Error marshaling JSON: %v", err)
			return
		}
		fmt.Print(string(data))
	}
}
