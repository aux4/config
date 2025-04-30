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
	Config map[string]interface{} `json:"config" yaml:"config"`
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
	value, found := getNestedValue(aux4Config.Config, path)
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
	setNestedValue(aux4Config.Config, path, value)

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
		value, found = getNestedValue(aux4Config.Config, path)
		if !found {
			value = make(map[string]interface{})
		}
	}

	mergedConfig, err := mergeConfig(value.(map[string]interface{}))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error merging config: %v\n", err)
		os.Exit(1)
	}

	if path == "" {
		aux4Config.Config = mergedConfig
	} else {
		setNestedValue(aux4Config.Config, path, mergedConfig)
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
		err = yaml.Unmarshal(data, &auxConfig)
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

func printValue(configValue interface{}) {
	switch value := configValue.(type) {
	case map[string]interface{}:
		data, _ := json.Marshal(value)
		fmt.Println(string(data))
	case Aux4Config:
		data, _ := json.Marshal(value.Config)
		fmt.Println(string(data))
	default:
		fmt.Println(value)
	}
}
