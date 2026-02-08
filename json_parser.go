package main

import (
	"encoding/json"
	"fmt"
)

func parseJSONWithOrder(data []byte) (any, error) {
	var rootMap OrderedMap
	if err := json.Unmarshal(data, &rootMap); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	if configValue, exists := rootMap.Get("config"); exists {
		return configValue, nil
	}

	return nil, fmt.Errorf("config section not found")
}
