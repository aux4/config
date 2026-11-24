package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// specVariable is the subset of command variable metadata that can be
// represented by a configuration schema.
type specVariable struct {
	Name     string
	Text     string
	Type     string
	Default  *string
	Hide     bool
	Private  bool
	Multiple bool
	Options  []string
}

// mergeSpec adds command variables to a structural config.spec.yaml. Existing
// fields are retained, including comments and metadata added by an author.
// The file is only replaced after every variable has been checked, so a
// conflict never leaves a partially generated spec behind.
func mergeSpec(filename, prefix string, variables []specVariable) error {
	if filename == "" {
		filename = "config.spec.yaml"
	}
	lockPath := filename + ".lock"
	unlock, err := acquireSpecLock(lockPath)
	if err != nil {
		return err
	}
	defer unlock()
	mode := os.FileMode(0644)
	var original []byte
	if info, statErr := os.Stat(filename); statErr == nil {
		mode = info.Mode().Perm()
		original, _ = os.ReadFile(filename)
	}

	root, err := readSpecDocument(filename)
	if err != nil {
		return err
	}
	if err := ensureSpecRoot(root); err != nil {
		return err
	}
	properties := mappingValue(root.Content[0], "properties")
	for _, variable := range variables {
		if variable.Private || strings.TrimSpace(variable.Name) == "" {
			continue
		}
		path := joinSpecPath(prefix, variable.Name)
		if err := mergeField(properties, path, variableNode(variable), path); err != nil {
			return err
		}
	}

	data, err := yaml.Marshal(root)
	if err != nil {
		return fmt.Errorf("encode config spec: %w", err)
	}
	if len(original) > 0 && string(original) == string(data) {
		return nil
	}
	dir := filepath.Dir(filename)
	tmp, err := os.CreateTemp(dir, ".config.spec.yaml-*")
	if err != nil {
		return fmt.Errorf("create spec temporary file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(mode); err == nil {
		_, err = tmp.Write(data)
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("write config spec: %w", err)
	}
	if err := os.Rename(tmpName, filename); err != nil {
		return fmt.Errorf("replace config spec: %w", err)
	}
	return nil
}

func acquireSpecLock(path string) (func(), error) {
	for attempt := 0; attempt < 300; attempt++ {
		lock, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err == nil {
			_ = lock.Close()
			return func() { _ = os.Remove(path) }, nil
		}
		if !os.IsExist(err) {
			return nil, fmt.Errorf("open spec lock: %w", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil, fmt.Errorf("timed out waiting for spec lock")
}

func readSpecDocument(filename string) (*yaml.Node, error) {
	data, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		return &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode}}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config spec: %w", err)
	}
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("parse config spec: %w", err)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err == nil {
		return nil, fmt.Errorf("config spec must contain one YAML document")
	} else if err != io.EOF {
		return nil, fmt.Errorf("parse config spec: %w", err)
	}
	if document.Kind != yaml.DocumentNode || len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("config spec root must be an object")
	}
	return &document, nil
}

func ensureSpecRoot(document *yaml.Node) error {
	root := document.Content[0]
	if keyNode(root, "type") == nil {
		appendMapping(root, scalar("type"), scalar("object"))
	} else if value := mappingValue(root, "type"); value.Value != "object" {
		return fmt.Errorf("config spec root type must be object")
	}
	properties := keyNode(root, "properties")
	if properties == nil {
		appendMapping(root, scalar("properties"), &yaml.Node{Kind: yaml.MappingNode})
	} else if properties.Kind != yaml.MappingNode {
		return fmt.Errorf("config spec properties must be an object")
	}
	return nil
}

func variableNode(variable specVariable) *yaml.Node {
	field := &yaml.Node{Kind: yaml.MappingNode}
	typ := strings.ToLower(variable.Type)
	if typ == "" {
		typ = "string"
	}
	if variable.Multiple {
		appendMapping(field, scalar("type"), scalar("array"))
		items := &yaml.Node{Kind: yaml.MappingNode}
		appendMapping(items, scalar("type"), scalar(typ))
		appendMapping(field, scalar("items"), items)
	} else {
		appendMapping(field, scalar("type"), scalar(typ))
	}
	if variable.Text != "" {
		appendMapping(field, scalar("description"), scalar(variable.Text))
	}
	if variable.Hide {
		appendMapping(field, scalar("hide"), typedScalar("true", "boolean"))
	}
	if len(variable.Options) > 0 {
		options := &yaml.Node{Kind: yaml.SequenceNode}
		for _, option := range variable.Options {
			options.Content = append(options.Content, scalar(option))
		}
		appendMapping(field, scalar("options"), options)
	}
	if variable.Default != nil {
		appendMapping(field, scalar("default"), typedScalar(*variable.Default, typ))
	}
	return field
}

func mergeField(properties *yaml.Node, path string, incoming *yaml.Node, fullPath string) error {
	parts := strings.Split(path, ".")
	name := parts[0]
	existing := keyNode(properties, name)
	if len(parts) > 1 {
		if existing == nil {
			existing = &yaml.Node{Kind: yaml.MappingNode}
			appendMapping(properties, scalar(name), existing)
		} else if existing.Kind != yaml.MappingNode {
			return fmt.Errorf("config spec conflict at %s", fullPath)
		} else if value := mappingValue(existing, "type"); value != nil && value.Value != "object" {
			return fmt.Errorf("config spec conflict at %s", fullPath)
		}
		if keyNode(existing, "type") == nil {
			appendMapping(existing, scalar("type"), scalar("object"))
		}
		if mappingValue(existing, "type").Value != "object" {
			return fmt.Errorf("config spec conflict at %s", fullPath)
		}
		child := keyNode(existing, "properties")
		if child == nil {
			child = &yaml.Node{Kind: yaml.MappingNode}
			appendMapping(existing, scalar("properties"), child)
		} else if child.Kind != yaml.MappingNode {
			return fmt.Errorf("config spec conflict at %s", fullPath)
		}
		return mergeField(child, strings.Join(parts[1:], "."), incoming, fullPath)
	}
	if existing == nil {
		appendMapping(properties, scalar(name), incoming)
		return nil
	}
	return mergeNode(existing, incoming, fullPath)
}

// mergeNode compares only keys supplied by the generated definition. This
// preserves author metadata such as titles, required flags, and comments.
func mergeNode(existing, incoming *yaml.Node, path string) error {
	if existing.Kind != incoming.Kind || (existing.Kind == yaml.ScalarNode && existing.Tag != incoming.Tag) {
		return fmt.Errorf("config spec conflict at %s", path)
	}
	if incoming.Kind == yaml.SequenceNode {
		if len(existing.Content) != len(incoming.Content) {
			return fmt.Errorf("config spec conflict at %s", path)
		}
		for i := range incoming.Content {
			if err := mergeNode(existing.Content[i], incoming.Content[i], path); err != nil {
				return err
			}
		}
		return nil
	}
	if incoming.Kind != yaml.MappingNode {
		if existing.Value != incoming.Value {
			return fmt.Errorf("config spec conflict at %s", path)
		}
		return nil
	}
	for i := 0; i+1 < len(incoming.Content); i += 2 {
		key := incoming.Content[i].Value
		value := incoming.Content[i+1]
		current := keyNode(existing, key)
		if current == nil {
			appendMapping(existing, incoming.Content[i], value)
			continue
		}
		if err := mergeNode(current, value, path+"."+key); err != nil {
			return err
		}
	}
	return nil
}

func joinSpecPath(prefix, name string) string {
	return strings.Trim(strings.TrimSpace(prefix)+"."+strings.TrimSpace(name), ".")
}

func scalar(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
}

func typedScalar(value, typ string) *yaml.Node {
	node := scalar(value)
	switch typ {
	case "boolean", "bool":
		node.Tag = "!!bool"
	case "integer", "int":
		node.Tag = "!!int"
	case "number", "float":
		node.Tag = "!!float"
	case "null":
		node.Tag = "!!null"
	}
	return node
}

func mappingValue(mapping *yaml.Node, key string) *yaml.Node {
	return keyNode(mapping, key)
}

func keyNode(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

func appendMapping(mapping, key, value *yaml.Node) {
	mapping.Content = append(mapping.Content, key, value)
}

func nodesEqual(left, right *yaml.Node) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Kind == right.Kind && left.Tag == right.Tag && left.Value == right.Value && len(left.Content) == len(right.Content) && equalContent(left.Content, right.Content)
}

func equalContent(left, right []*yaml.Node) bool {
	for i := range left {
		if !nodesEqual(left[i], right[i]) {
			return false
		}
	}
	return true
}
