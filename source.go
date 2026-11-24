package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// commandVariables reads the public variables advertised by an aux4 command.
// The command is invoked with --help --json, which asks aux4 to inspect the
// command without running its execute instructions.
func commandVariables(command []string) ([]specVariable, error) {
	if len(command) == 0 {
		return nil, fmt.Errorf("command is required")
	}
	for _, part := range command {
		if part == "" || strings.HasPrefix(part, "-") {
			return nil, fmt.Errorf("command path contains invalid action %q", part)
		}
	}

	args := append([]string{}, command...)
	args = append(args, "--help", "--json")
	cmd := exec.Command("aux4", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			return nil, fmt.Errorf("read command help: %s", bytes.TrimSpace(exitErr.Stderr))
		}
		return nil, fmt.Errorf("read command help: %w", err)
	}

	variables, err := parseCommandVariables(output)
	if err != nil {
		return nil, err
	}
	return enrichCommandVariableTypes(command, variables), nil
}

// The public help JSON intentionally omits type metadata. Read the package
// manifest identified by `aux4 aux4 which` to retain declared types.
func enrichCommandVariableTypes(command []string, variables []specVariable) []specVariable {
	manifest, profile, commandName, err := commandManifest(command)
	if err != nil {
		return variables
	}
	var pack struct {
		Profiles []struct {
			Name     string `json:"name"`
			Commands []struct {
				Name string `json:"name"`
				Help *struct {
					Variables []struct {
						Name string `json:"name"`
						Type string `json:"type"`
					} `json:"variables"`
				} `json:"help"`
			} `json:"commands"`
		} `json:"profiles"`
	}
	if json.Unmarshal(manifest, &pack) != nil {
		return variables
	}
	types := map[string]string{}
	for _, p := range pack.Profiles {
		if p.Name != profile {
			continue
		}
		for _, c := range p.Commands {
			if c.Name != commandName || c.Help == nil {
				continue
			}
			for _, v := range c.Help.Variables {
				types[v.Name] = v.Type
			}
		}
	}
	for i := range variables {
		if variables[i].Type == "" {
			variables[i].Type = types[variables[i].Name]
		}
	}
	return variables
}

func commandManifest(command []string) ([]byte, string, string, error) {
	args := append([]string{"aux4", "aux4", "which"}, command...)
	output, err := exec.Command("aux4", args...).Output()
	if err != nil {
		return nil, "", "", err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 || lines[0] == "" {
		return nil, "", "", fmt.Errorf("invalid command location")
	}
	parts := strings.SplitN(lines[1], " → ", 2)
	if len(parts) != 2 {
		return nil, "", "", fmt.Errorf("invalid command location")
	}
	profile := strings.TrimSpace(strings.Fields(parts[0])[len(strings.Fields(parts[0]))-1])
	data, err := os.ReadFile(lines[0])
	return data, profile, strings.TrimSpace(parts[1]), err
}

type commandHelpJSON struct {
	Params    []commandVariableJSON `json:"params"`
	Variables []commandVariableJSON `json:"variables"`
}

type commandVariableJSON struct {
	Name     string          `json:"name"`
	Text     string          `json:"text"`
	Default  json.RawMessage `json:"default"`
	Type     string          `json:"type"`
	Hide     bool            `json:"hide"`
	Private  bool            `json:"private"`
	Multiple bool            `json:"multiple"`
	Options  []string        `json:"options"`
}

func parseCommandVariables(data []byte) ([]specVariable, error) {
	var help commandHelpJSON
	if err := json.Unmarshal(data, &help); err != nil {
		return nil, fmt.Errorf("parse command help: %w", err)
	}
	params := help.Params
	if params == nil {
		params = help.Variables
	}

	variables := make([]specVariable, 0, len(params))
	for _, param := range params {
		if param.Name == "" || param.Private {
			continue
		}
		variable := specVariable{
			Name:     param.Name,
			Text:     param.Text,
			Type:     param.Type,
			Hide:     param.Hide,
			Multiple: param.Multiple,
			Options:  append([]string(nil), param.Options...),
		}
		if len(param.Default) > 0 && string(param.Default) != "null" {
			var value any
			if err := json.Unmarshal(param.Default, &value); err != nil {
				return nil, fmt.Errorf("parse default for %q: %w", param.Name, err)
			}
			text, err := defaultText(value)
			if err != nil {
				return nil, fmt.Errorf("parse default for %q: %w", param.Name, err)
			}
			variable.Default = &text
		}
		variables = append(variables, variable)
	}
	return variables, nil
}

func defaultText(value any) (string, error) {
	if text, ok := value.(string); ok {
		return text, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
