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

	// Convert merged config back to OrderedMap to preserve order
	var mergedOrderedMap interface{}
	if path == "" {
		mergedOrderedMap = convertMapToOrderedMap(mergedConfig)
		aux4Config.Config = mergedOrderedMap
	} else {
		mergedOrderedMap = convertMapToOrderedMap(mergedConfig)
		aux4Config.Config = setNestedValueInOrderedMap(aux4Config.Config, path, mergedOrderedMap)
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
		// Convert OrderedMap back to regular maps for proper JSON serialization
		configForSave := Aux4Config{Config: convertOrderedMapToMap(auxConfig.Config)}
		data, err = json.Marshal(configForSave)
	case ".yaml", ".yml":
		// Convert OrderedMap back to regular maps for proper YAML serialization
		configForSave := Aux4Config{Config: convertOrderedMapToMap(auxConfig.Config)}
		var buf strings.Builder
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		err = encoder.Encode(configForSave)
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

func convertMapToOrderedMap(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		// For regular maps (from merge operations), we can't preserve original order
		// since Go maps have random iteration. We'll just use the iteration order.
		om := newOrderedMap()
		for key, val := range v {
			om.Set(key, convertMapToOrderedMap(val))
		}
		return om
	case *OrderedMap:
		// OrderedMap is already in the right format, but recursively process values
		om := newOrderedMap()
		for _, key := range v.Keys {
			om.Set(key, convertMapToOrderedMap(v.Values[key]))
		}
		return om
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = convertMapToOrderedMap(item)
		}
		return result
	default:
		return v
	}
}

func convertOrderedMapToMap(value interface{}) interface{} {
	switch v := value.(type) {
	case *OrderedMap:
		result := make(map[string]interface{})
		for _, key := range v.Keys {
			result[key] = convertOrderedMapToMap(v.Values[key])
		}
		return result
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			result[key] = convertOrderedMapToMap(val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = convertOrderedMapToMap(val)
		}
		return result
	default:
		return v
	}
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
		mapConfig = make(map[string]interface{})
		for key, val := range v.Values {
			mapConfig[key] = convertOrderedMapToMap(val)
		}
	case map[string]interface{}:
		mapConfig = v
	default:
		mapConfig = make(map[string]interface{})
	}

	setNestedValue(mapConfig, path, value)

	// Convert back to OrderedMap if it was originally
	if origOm, ok := config.(*OrderedMap); ok {
		result := newOrderedMap()
		// First, preserve original order for existing keys
		for _, key := range origOm.Keys {
			if val, exists := mapConfig[key]; exists {
				result.Set(key, val)
			}
		}
		// Then add any new keys that weren't in the original
		for key, val := range mapConfig {
			if _, exists := result.Get(key); !exists {
				result.Set(key, val)
			}
		}
		return result
	}

	return mapConfig
}

func convertBackToOrderedMap(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		// Check if this was originally an OrderedMap by looking for nested structures
		result := newOrderedMap()
		for key, val := range v {
			result.Set(key, convertBackToOrderedMap(val))
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = convertBackToOrderedMap(val)
		}
		return result
	default:
		return v
	}
}

func mergeConfig(root map[string]interface{}) (map[string]interface{}, error) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return map[string]interface{}{}, err
	}

	var raw map[string]interface{}

	// Use the same order-preserving JSON parsing as loadConfig
	rawOrderedMap, err := parseJSONWithOrder(input)
	if err != nil {
		// If JSON parsing fails, fall back to regular JSON parsing
		if err := json.Unmarshal(input, &raw); err != nil {
			return root, err
		}
	} else {
		// Convert OrderedMap to regular map for merging
		raw = convertOrderedMapToMap(rawOrderedMap).(map[string]interface{})
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

func parseJSONWithOrder(data []byte) (interface{}, error) {
	// Parse the JSON while preserving key order using a custom decoder
	rootValue, err := parseJSONValue(string(data))
	if err != nil {
		return nil, err
	}

	// Extract the config section
	if rootMap, ok := rootValue.(*OrderedMap); ok {
		if configValue, exists := rootMap.Get("config"); exists {
			return configValue, nil
		}
	}

	return nil, fmt.Errorf("config section not found")
}

func parseJSONValue(jsonStr string) (interface{}, error) {
	jsonStr = strings.TrimSpace(jsonStr)

	if len(jsonStr) == 0 {
		return nil, fmt.Errorf("empty JSON")
	}

	switch jsonStr[0] {
	case '{':
		return parseJSONObject(jsonStr)
	case '[':
		return parseJSONArray(jsonStr)
	case '"':
		return parseJSONString(jsonStr)
	case 't':
		if strings.HasPrefix(jsonStr, "true") {
			return true, nil
		}
		return nil, fmt.Errorf("invalid JSON value: %s", jsonStr)
	case 'f':
		if strings.HasPrefix(jsonStr, "false") {
			return false, nil
		}
		return nil, fmt.Errorf("invalid JSON value: %s", jsonStr)
	case 'n':
		if strings.HasPrefix(jsonStr, "null") {
			return nil, nil
		}
		return nil, fmt.Errorf("invalid JSON value: %s", jsonStr)
	default:
		// Try to parse as number
		return parseJSONNumber(jsonStr)
	}
}

func parseJSONObject(jsonStr string) (*OrderedMap, error) {
	if !strings.HasPrefix(jsonStr, "{") || !strings.HasSuffix(jsonStr, "}") {
		return nil, fmt.Errorf("invalid JSON object")
	}

	om := newOrderedMap()
	content := strings.TrimSpace(jsonStr[1 : len(jsonStr)-1])

	if content == "" {
		return om, nil
	}

	// Parse key-value pairs while preserving order
	i := 0
	for i < len(content) {
		// Skip whitespace
		for i < len(content) && (content[i] == ' ' || content[i] == '\t' || content[i] == '\n' || content[i] == '\r') {
			i++
		}
		if i >= len(content) {
			break
		}

		// Parse key
		if content[i] != '"' {
			return nil, fmt.Errorf("expected quoted key")
		}

		keyEnd := i + 1
		for keyEnd < len(content) && content[keyEnd] != '"' {
			if content[keyEnd] == '\\' {
				keyEnd++ // Skip escaped character
			}
			keyEnd++
		}
		if keyEnd >= len(content) {
			return nil, fmt.Errorf("unterminated string")
		}

		key := content[i+1 : keyEnd]
		i = keyEnd + 1

		// Skip whitespace and colon
		for i < len(content) && (content[i] == ' ' || content[i] == '\t' || content[i] == '\n' || content[i] == '\r') {
			i++
		}
		if i >= len(content) || content[i] != ':' {
			return nil, fmt.Errorf("expected colon after key")
		}
		i++

		// Find the value
		for i < len(content) && (content[i] == ' ' || content[i] == '\t' || content[i] == '\n' || content[i] == '\r') {
			i++
		}

		valueStart := i
		valueEnd := findJSONValueEnd(content, i)
		if valueEnd == -1 {
			return nil, fmt.Errorf("invalid value")
		}

		valueStr := content[valueStart:valueEnd]
		value, err := parseJSONValue(valueStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing value for key %s: %v", key, err)
		}

		om.Set(key, value)
		i = valueEnd

		// Skip whitespace
		for i < len(content) && (content[i] == ' ' || content[i] == '\t' || content[i] == '\n' || content[i] == '\r') {
			i++
		}

		// Check for comma or end
		if i < len(content) {
			if content[i] == ',' {
				i++
			} else if i < len(content) {
				return nil, fmt.Errorf("expected comma or end of object")
			}
		}
	}

	return om, nil
}

func parseJSONArray(jsonStr string) ([]interface{}, error) {
	if !strings.HasPrefix(jsonStr, "[") || !strings.HasSuffix(jsonStr, "]") {
		return nil, fmt.Errorf("invalid JSON array")
	}

	content := strings.TrimSpace(jsonStr[1 : len(jsonStr)-1])
	if content == "" {
		return []interface{}{}, nil
	}

	var result []interface{}
	i := 0

	for i < len(content) {
		// Skip whitespace
		for i < len(content) && (content[i] == ' ' || content[i] == '\t' || content[i] == '\n' || content[i] == '\r') {
			i++
		}
		if i >= len(content) {
			break
		}

		valueEnd := findJSONValueEnd(content, i)
		if valueEnd == -1 {
			return nil, fmt.Errorf("invalid array value")
		}

		valueStr := content[i:valueEnd]
		value, err := parseJSONValue(valueStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing array value: %v", err)
		}

		result = append(result, value)
		i = valueEnd

		// Skip whitespace
		for i < len(content) && (content[i] == ' ' || content[i] == '\t' || content[i] == '\n' || content[i] == '\r') {
			i++
		}

		// Check for comma
		if i < len(content) && content[i] == ',' {
			i++
		}
	}

	return result, nil
}

func parseJSONString(jsonStr string) (string, error) {
	if len(jsonStr) < 2 || !strings.HasPrefix(jsonStr, "\"") || !strings.HasSuffix(jsonStr, "\"") {
		return "", fmt.Errorf("invalid JSON string")
	}

	// Simple unescaping - for a full implementation you'd need to handle all escape sequences
	content := jsonStr[1 : len(jsonStr)-1]
	content = strings.ReplaceAll(content, "\\\"", "\"")
	content = strings.ReplaceAll(content, "\\\\", "\\")
	content = strings.ReplaceAll(content, "\\n", "\n")
	content = strings.ReplaceAll(content, "\\r", "\r")
	content = strings.ReplaceAll(content, "\\t", "\t")

	return content, nil
}

func parseJSONNumber(jsonStr string) (interface{}, error) {
	jsonStr = strings.TrimSpace(jsonStr)

	// Try to parse as integer first
	if strings.Contains(jsonStr, ".") {
		// Parse as float
		var f float64
		err := json.Unmarshal([]byte(jsonStr), &f)
		return f, err
	} else {
		// Parse as integer
		var i int
		err := json.Unmarshal([]byte(jsonStr), &i)
		return i, err
	}
}

func findJSONValueEnd(content string, start int) int {
	if start >= len(content) {
		return -1
	}

	switch content[start] {
	case '"':
		// Find end of string
		i := start + 1
		for i < len(content) {
			if content[i] == '"' && (i == start+1 || content[i-1] != '\\') {
				return i + 1
			}
			if content[i] == '\\' {
				i++ // Skip escaped character
			}
			i++
		}
		return -1
	case '{':
		// Find end of object
		depth := 0
		for i := start; i < len(content); i++ {
			if content[i] == '{' {
				depth++
			} else if content[i] == '}' {
				depth--
				if depth == 0 {
					return i + 1
				}
			}
		}
		return -1
	case '[':
		// Find end of array
		depth := 0
		for i := start; i < len(content); i++ {
			if content[i] == '[' {
				depth++
			} else if content[i] == ']' {
				depth--
				if depth == 0 {
					return i + 1
				}
			}
		}
		return -1
	default:
		// Find end of primitive value (number, boolean, null)
		i := start
		for i < len(content) && content[i] != ',' && content[i] != '}' && content[i] != ']' {
			i++
		}
		// Trim trailing whitespace
		for i > start && (content[i-1] == ' ' || content[i-1] == '\t' || content[i-1] == '\n' || content[i-1] == '\r') {
			i--
		}
		return i
	}
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
	printValueWithContext(configValue, false)
}

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
	case OrderedMap:
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
		// For regular maps, we need to handle nested OrderedMaps properly
		// Convert any nested OrderedMaps to regular maps first, then marshal
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
			// Standalone strings should not have quotes
			fmt.Print(value)
		}
	default:
		data, _ := json.Marshal(value)
		if string(data) == "null" {
			fmt.Print(value)
		} else {
			fmt.Print(string(data))
		}
	}
}
