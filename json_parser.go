package main

import (
	"encoding/json"
	"fmt"
)

// parseJSONWithOrder parses JSON while preserving key order using standard json.Unmarshal
// with our custom OrderedMap unmarshaler
func parseJSONWithOrder(data []byte) (any, error) {
	var rootMap OrderedMap
	if err := json.Unmarshal(data, &rootMap); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Extract the config section
	if configValue, exists := rootMap.Get("config"); exists {
		return configValue, nil
	}

	return nil, fmt.Errorf("config section not found")
}