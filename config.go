package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Aux4Config struct {
	Config any `json:"config" yaml:"config"`
}

func findConfigFile() string {
	files := []string{"config.yaml", "config.yml", "config.json"}
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			return file
		}
	}
	return ""
}

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
		var node yaml.Node
		if err := yaml.Unmarshal(data, &node); err != nil {
			return Aux4Config{}, err
		}

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

func saveConfig(filename string, auxConfig Aux4Config) error {
	switch filepath.Ext(filename) {
	case ".json":
		config := make(map[string]any)
		config["config"] = convertOrderedMapToMap(auxConfig.Config)

		data, err := json.Marshal(config)
		if err != nil {
			return err
		}
		return os.WriteFile(filename, data, 0644)
	case ".yaml", ".yml":
		configForSave := Aux4Config{Config: convertOrderedMapToMap(auxConfig.Config)}
		var buf strings.Builder
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		err := encoder.Encode(configForSave)
		encoder.Close()
		if err != nil {
			return err
		}
		data := []byte(buf.String())
		return os.WriteFile(filename, data, 0644)
	default:
		return fmt.Errorf("unsupported file format")
	}
}

func convertToConfig(property map[string]any) map[string]any {
	config := make(map[string]any)
	for key, value := range property {
		if nestedProperty, ok := value.(map[string]any); ok {
			config[key] = convertToConfig(nestedProperty)
		} else {
			config[key] = value
		}
	}
	return config
}
