package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// parseJSONWithOrder parses JSON while preserving key order
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

// parseJSONValue parses a JSON value while preserving order
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
	case 't', 'f':
		return parseJSONBoolean(jsonStr)
	case 'n':
		return parseJSONNull(jsonStr)
	default:
		return parseJSONNumber(jsonStr)
	}
}

// parseJSONObject parses a JSON object while preserving key order
func parseJSONObject(jsonStr string) (*OrderedMap, error) {
	jsonStr = strings.TrimSpace(jsonStr)
	if !strings.HasPrefix(jsonStr, "{") || !strings.HasSuffix(jsonStr, "}") {
		return nil, fmt.Errorf("invalid JSON object")
	}

	content := jsonStr[1 : len(jsonStr)-1]
	content = strings.TrimSpace(content)

	result := newOrderedMap()

	if content == "" {
		return result, nil
	}

	pos := 0
	for pos < len(content) {
		// Skip whitespace
		for pos < len(content) && unicode.IsSpace(rune(content[pos])) {
			pos++
		}

		if pos >= len(content) {
			break
		}

		// Parse key
		if content[pos] != '"' {
			return nil, fmt.Errorf("expected quoted key at position %d", pos)
		}

		keyEnd := pos + 1
		for keyEnd < len(content) {
			if content[keyEnd] == '"' && (keyEnd == pos+1 || content[keyEnd-1] != '\\') {
				break
			}
			keyEnd++
		}

		if keyEnd >= len(content) {
			return nil, fmt.Errorf("unterminated key string")
		}

		key := content[pos+1 : keyEnd]
		pos = keyEnd + 1

		// Skip whitespace after key
		for pos < len(content) && unicode.IsSpace(rune(content[pos])) {
			pos++
		}

		// Expect colon
		if pos >= len(content) || content[pos] != ':' {
			return nil, fmt.Errorf("expected ':' after key")
		}
		pos++

		// Skip whitespace after colon
		for pos < len(content) && unicode.IsSpace(rune(content[pos])) {
			pos++
		}

		// Find value end
		valueStart := pos
		valueEnd := findJSONValueEnd(content, pos)
		if valueEnd == -1 {
			return nil, fmt.Errorf("could not find value end")
		}

		valueStr := content[valueStart:valueEnd]
		value, err := parseJSONValue(valueStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing value for key '%s': %v", key, err)
		}

		result.Set(key, value)
		pos = valueEnd

		// Skip whitespace
		for pos < len(content) && unicode.IsSpace(rune(content[pos])) {
			pos++
		}

		// Check for comma or end
		if pos < len(content) {
			if content[pos] == ',' {
				pos++
			} else if content[pos] != '}' {
				// Continue parsing if we're not at the end
			}
		}
	}

	return result, nil
}

// parseJSONArray parses a JSON array
func parseJSONArray(jsonStr string) ([]interface{}, error) {
	jsonStr = strings.TrimSpace(jsonStr)
	if !strings.HasPrefix(jsonStr, "[") || !strings.HasSuffix(jsonStr, "]") {
		return nil, fmt.Errorf("invalid JSON array")
	}

	content := jsonStr[1 : len(jsonStr)-1]
	content = strings.TrimSpace(content)

	if content == "" {
		return []interface{}{}, nil
	}

	var result []interface{}
	pos := 0

	for pos < len(content) {
		// Skip whitespace
		for pos < len(content) && unicode.IsSpace(rune(content[pos])) {
			pos++
		}

		if pos >= len(content) {
			break
		}

		// Find value end
		valueStart := pos
		valueEnd := findJSONValueEnd(content, pos)
		if valueEnd == -1 {
			return nil, fmt.Errorf("could not find array element end")
		}

		valueStr := content[valueStart:valueEnd]
		value, err := parseJSONValue(valueStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing array element: %v", err)
		}

		result = append(result, value)
		pos = valueEnd

		// Skip whitespace
		for pos < len(content) && unicode.IsSpace(rune(content[pos])) {
			pos++
		}

		// Check for comma
		if pos < len(content) && content[pos] == ',' {
			pos++
		}
	}

	return result, nil
}

// findJSONValueEnd finds the end of a JSON value
func findJSONValueEnd(content string, start int) int {
	if start >= len(content) {
		return -1
	}

	switch content[start] {
	case '"':
		// String value
		pos := start + 1
		for pos < len(content) {
			if content[pos] == '"' && (pos == start+1 || content[pos-1] != '\\') {
				return pos + 1
			}
			pos++
		}
		return -1
	case '{':
		// Object value
		braceCount := 0
		inString := false
		for pos := start; pos < len(content); pos++ {
			if content[pos] == '"' && (pos == 0 || content[pos-1] != '\\') {
				inString = !inString
			} else if !inString {
				if content[pos] == '{' {
					braceCount++
				} else if content[pos] == '}' {
					braceCount--
					if braceCount == 0 {
						return pos + 1
					}
				}
			}
		}
		return -1
	case '[':
		// Array value
		bracketCount := 0
		inString := false
		for pos := start; pos < len(content); pos++ {
			if content[pos] == '"' && (pos == 0 || content[pos-1] != '\\') {
				inString = !inString
			} else if !inString {
				if content[pos] == '[' {
					bracketCount++
				} else if content[pos] == ']' {
					bracketCount--
					if bracketCount == 0 {
						return pos + 1
					}
				}
			}
		}
		return -1
	default:
		// Primitive value (number, boolean, null)
		pos := start
		for pos < len(content) && !unicode.IsSpace(rune(content[pos])) && content[pos] != ',' && content[pos] != '}' && content[pos] != ']' {
			pos++
		}
		return pos
	}
}

// parseJSONString parses a JSON string
func parseJSONString(jsonStr string) (string, error) {
	if len(jsonStr) < 2 || !strings.HasPrefix(jsonStr, "\"") || !strings.HasSuffix(jsonStr, "\"") {
		return "", fmt.Errorf("invalid JSON string")
	}
	return jsonStr[1 : len(jsonStr)-1], nil
}

// parseJSONNumber parses a JSON number
func parseJSONNumber(jsonStr string) (interface{}, error) {
	if strings.Contains(jsonStr, ".") {
		return strconv.ParseFloat(jsonStr, 64)
	}
	return strconv.ParseInt(jsonStr, 10, 64)
}

// parseJSONBoolean parses a JSON boolean
func parseJSONBoolean(jsonStr string) (bool, error) {
	if jsonStr == "true" {
		return true, nil
	} else if jsonStr == "false" {
		return false, nil
	}
	return false, fmt.Errorf("invalid boolean value: %s", jsonStr)
}

// parseJSONNull parses a JSON null
func parseJSONNull(jsonStr string) (interface{}, error) {
	if jsonStr == "null" {
		return nil, nil
	}
	return nil, fmt.Errorf("invalid null value: %s", jsonStr)
}