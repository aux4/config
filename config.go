package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Aux4Config struct {
	Config interface{} `json:"config" yaml:"config"`
}

// findConfigFile finds the first config file in order of preference
func findConfigFile() string {
	files := []string{"config.json", "config.yaml", "config.yml"}
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			return file
		}
	}
	return ""
}

// loadConfig loads configuration from a file
func loadConfig(filename string) (Aux4Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Aux4Config{}, err
	}

	var auxConfig Aux4Config
	switch filepath.Ext(filename) {
	case ".json":
		config, err := parseJSONWithOrder(data)
		if err != nil {
			return Aux4Config{}, err
		}
		auxConfig = Aux4Config{Config: config}
	case ".yaml", ".yml":
		// Use yaml.Node to preserve order, then convert to ordered map
		var node yaml.Node
		if err := yaml.Unmarshal(data, &node); err != nil {
			return Aux4Config{}, err
		}

		// Find the config node
		var configNode *yaml.Node
		if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
			rootNode := node.Content[0]
			if rootNode.Kind == yaml.MappingNode {
				for i := 0; i < len(rootNode.Content); i += 2 {
					keyNode := rootNode.Content[i]
					valueNode := rootNode.Content[i+1]
					var key string
					if keyNode.Decode(&key) == nil && key == "config" {
						configNode = valueNode
						break
					}
				}
			}
		}

		if configNode != nil {
			config, err := nodeToOrderedMap(configNode)
			if err != nil {
				return Aux4Config{}, err
			}
			auxConfig = Aux4Config{Config: config}
		} else {
			auxConfig = Aux4Config{Config: newOrderedMap()}
		}
	default:
		return Aux4Config{}, fmt.Errorf("unsupported file format")
	}

	return auxConfig, nil
}

// saveConfig saves configuration to a file
func saveConfig(filename string, auxConfig Aux4Config) error {
	switch filepath.Ext(filename) {
	case ".json":
		// Convert OrderedMap to regular map for saving
		config := make(map[string]interface{})
		config["config"] = convertOrderedMapToMap(auxConfig.Config)

		data, err := json.Marshal(config)
		if err != nil {
			return err
		}
		return os.WriteFile(filename, data, 0644)
	case ".yaml", ".yml":
		// Convert OrderedMap to regular map for saving
		config := make(map[string]interface{})
		config["config"] = convertOrderedMapToMap(auxConfig.Config)

		data, err := yaml.Marshal(config)
		if err != nil {
			return err
		}
		return os.WriteFile(filename, data, 0644)
	default:
		return fmt.Errorf("unsupported file format")
	}
}

// convertToConfig wraps a property in a config structure
func convertToConfig(property map[string]interface{}) map[string]interface{} {
	config := make(map[string]interface{})
	for key, value := range property {
		if nestedProperty, ok := value.(map[string]interface{}); ok {
			config[key] = convertToConfig(nestedProperty)
		} else {
			config[key] = value
		}
	}
	return config
}