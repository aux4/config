package main

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func nodeToOrderedMap(node *yaml.Node) (*OrderedMap, error) {
	result := newOrderedMap()

	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			var key string
			if err := keyNode.Decode(&key); err != nil {
				return nil, fmt.Errorf("failed to decode key: %v", err)
			}

			value, err := convertYAMLNode(valueNode)
			if err != nil {
				return nil, fmt.Errorf("failed to convert value for key '%s': %v", key, err)
			}

			result.Set(key, value)
		}
	}

	return result, nil
}

func convertYAMLNode(node *yaml.Node) (any, error) {
	switch node.Kind {
	case yaml.MappingNode:
		return nodeToOrderedMap(node)
	case yaml.SequenceNode:
		var result []any
		for _, child := range node.Content {
			value, err := convertYAMLNode(child)
			if err != nil {
				return nil, err
			}
			result = append(result, value)
		}
		return result, nil
	case yaml.ScalarNode:
		var value any
		if err := node.Decode(&value); err != nil {
			return nil, err
		}
		return value, nil
	default:
		return nil, fmt.Errorf("unsupported YAML node kind: %v", node.Kind)
	}
}
