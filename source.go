package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

// commandVariables reads the public variables advertised by an aux4 command.
// The command is invoked with --help --json, which asks aux4 to inspect the
// command without running its execute instructions.
func commandVariables(command []string) ([]specVariable, error) {
	if len(command) == 0 {
		return nil, fmt.Errorf("command is required")
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

	return parseCommandVariables(output)
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
