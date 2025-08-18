package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Aux4Config struct {
	Config interface{} `json:"config" yaml:"config"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Invalid command")
		os.Exit(1)
	}

	action := os.Args[1]
	configFile := ""
	if len(os.Args) > 2 {
		configFile = os.Args[2]
	}

	if configFile == "" {
		configFile = findConfigFile()
	}

	if configFile == "" {
		fmt.Fprintln(os.Stderr, "No config file found")
		os.Exit(1)
	}

	aux4Config, err := loadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	switch action {
	case "get":
		getConfig(aux4Config)
	case "set":
		setConfig(aux4Config, configFile)
	case "merge":
		mergeAux4Config(aux4Config, configFile)
	default:
		fmt.Fprintln(os.Stderr, "Command not found")
		os.Exit(1)
	}
}

func getConfig(aux4Config Aux4Config) {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "Mising parameters")
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
		fmt.Fprintln(os.Stderr, "Mising parameters")
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
		fmt.Fprintln(os.Stderr, "Mising parameters")
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

func findConfigFile() string {
	for _, ext := range []string{"yaml", "yml", "json"} {
		path := "config." + ext
		if _, err := os.Stat(path); err == nil {
			return path
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
		var raw map[string]interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			return Aux4Config{}, err
		}
		auxConfig = Aux4Config{Config: convertToConfig(raw["config"].(map[string]interface{}))}
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
			// Fallback to regular unmarshal
			err = yaml.Unmarshal(data, &auxConfig)
		}
	}

	return auxConfig, err
}

func saveConfig(filename string, auxConfig Aux4Config) error {
	var data []byte
	var err error

	switch filepath.Ext(filename) {
	case ".json":
		data, err = json.MarshalIndent(auxConfig, "", "  ")
	case ".yaml", ".yml":
		var buf strings.Builder
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		err = encoder.Encode(auxConfig)
		encoder.Close()
		data = []byte(buf.String())
	}

	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

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

func getNestedValue(config map[string]interface{}, path string) (interface{}, bool) {
	if path == "" {
		return config, true
	}

	var value interface{} = config

	keys := strings.Split(path, "/")

	for _, key := range keys {
		if property, ok := value.(map[string]interface{}); ok {
			value, ok = property[key]
			if !ok {
				return nil, false
			}
		} else {
			return nil, false
		}
	}

	return value, true
}

func getNestedValueFromOrderedMap(value interface{}, path string) (interface{}, bool) {
	if path == "" {
		return value, true
	}

	keys := strings.Split(path, "/")
	current := value

	for _, key := range keys {
		switch v := current.(type) {
		case *OrderedMap:
			next, exists := v.Get(key)
			if !exists {
				return nil, false
			}
			current = next
		case map[string]interface{}:
			next, exists := v[key]
			if !exists {
				return nil, false
			}
			current = next
		default:
			return nil, false
		}
	}

	return current, true
}

func setNestedValue(config map[string]interface{}, path string, value any) {
	keys := strings.Split(path, "/")
	lastKey := keys[len(keys)-1]
	property := config

	for _, key := range keys[:len(keys)-1] {
		if _, ok := property[key].(map[string]interface{}); !ok {
			property[key] = make(map[string]interface{})
		}
		property = property[key].(map[string]interface{})
	}

	property[lastKey] = value
}

func setNestedValueInOrderedMap(config interface{}, path string, value any) interface{} {
	// Convert to map for modification, then back to OrderedMap
	var mapConfig map[string]interface{}
	
	switch v := config.(type) {
	case *OrderedMap:
		mapConfig = v.Values
	case map[string]interface{}:
		mapConfig = v
	default:
		mapConfig = make(map[string]interface{})
	}
	
	setNestedValue(mapConfig, path, value)
	
	// Convert back to OrderedMap if it was originally
	if _, ok := config.(*OrderedMap); ok {
		result := newOrderedMap()
		if origOm, ok := config.(*OrderedMap); ok {
			// Preserve order
			for _, key := range origOm.Keys {
				result.Set(key, mapConfig[key])
			}
		}
		return result
	}
	
	return mapConfig
}

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

func deepMerge(sourceConfig map[string]interface{}, incomingConfig map[string]interface{}) map[string]interface{} {
	for key, value := range incomingConfig {
		if nestedIncomingConfig, ok := value.(map[string]interface{}); ok {
			if nestedSourceConfig, ok := sourceConfig[key].(map[string]interface{}); ok {
				sourceConfig[key] = deepMerge(nestedSourceConfig, nestedIncomingConfig)
			} else {
				sourceConfig[key] = nestedIncomingConfig
			}
		} else {
			sourceConfig[key] = value
		}
	}

	return sourceConfig
}

// OrderedMap represents a map that preserves insertion order
type OrderedMap struct {
	Keys   []string
	Values map[string]interface{}
}

func newOrderedMap() *OrderedMap {
	return &OrderedMap{
		Keys:   make([]string, 0),
		Values: make(map[string]interface{}),
	}
}

func (om *OrderedMap) Set(key string, value interface{}) {
	if _, exists := om.Values[key]; !exists {
		om.Keys = append(om.Keys, key)
	}
	om.Values[key] = value
}

func (om *OrderedMap) Get(key string) (interface{}, bool) {
	value, exists := om.Values[key]
	return value, exists
}

func nodeToOrderedMap(node *yaml.Node) (interface{}, error) {
	switch node.Kind {
	case yaml.MappingNode:
		om := newOrderedMap()
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			
			var key string
			if err := keyNode.Decode(&key); err != nil {
				continue
			}
			
			value, err := nodeToOrderedMap(valueNode)
			if err != nil {
				continue
			}
			om.Set(key, value)
		}
		return om, nil
	case yaml.SequenceNode:
		var result []interface{}
		for _, childNode := range node.Content {
			value, err := nodeToOrderedMap(childNode)
			if err != nil {
				continue
			}
			result = append(result, value)
		}
		return result, nil
	case yaml.ScalarNode:
		var value interface{}
		err := node.Decode(&value)
		return value, err
	}
	return nil, fmt.Errorf("unsupported node kind: %v", node.Kind)
}

func printValue(configValue interface{}) {
	switch value := configValue.(type) {
	case *OrderedMap:
		fmt.Print("{")
		for i, key := range value.Keys {
			if i > 0 {
				fmt.Print(",")
			}
			keyJSON, _ := json.Marshal(key)
			fmt.Printf("%s:", keyJSON)
			printValue(value.Values[key])
		}
		fmt.Print("}")
	case map[string]interface{}:
		// Convert regular map to OrderedMap if possible, otherwise use regular JSON
		data, _ := json.Marshal(value)
		fmt.Print(string(data))
	case Aux4Config:
		printValue(value.Config)
	default:
		data, _ := json.Marshal(value)
		if string(data) == "null" {
			fmt.Print(value)
		} else {
			fmt.Print(string(data))
		}
	}
}
