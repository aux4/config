package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Inspect each routing step before extending the path. aux4 treats extra
// positional arguments on a leaf as input, even when help was requested.
func commandVariables(command []string) ([]specVariable, error) {
	if len(command) == 0 {
		return nil, fmt.Errorf("command is required")
	}
	for i, part := range command {
		if part == "" || strings.HasPrefix(part, "-") || strings.ContainsAny(part, "\r\n") {
			return nil, fmt.Errorf("invalid command name %q", part)
		}
		if i == len(command)-1 {
			break
		}
		src, err := inspectCommand(command[:i+1], "--showSource")
		if err != nil {
			return nil, err
		}
		if !staticProfile.Match(bytes.TrimSpace(src)) {
			return nil, fmt.Errorf("%s does not route to a static command profile", strings.Join(command[:i+1], " "))
		}
	}
	location, err := inspectCommand(command, "--whereIsIt")
	if err != nil {
		return nil, err
	}
	header, path, _ := strings.Cut(strings.TrimSpace(string(location)), "\n")
	owner, name, found := strings.Cut(header, " → ")
	parts := strings.Fields(owner)
	if !found || len(parts) == 0 || name != command[len(command)-1] {
		return nil, fmt.Errorf("unexpected aux4 command location")
	}
	if path == "" {
		// Built-in commands have no manifest.
		help, err := inspectCommand(command, "--help", "--json")
		if err != nil {
			return nil, err
		}
		var result struct{ Params []specVariable }
		if err := json.Unmarshal(help, &result); err != nil {
			return nil, fmt.Errorf("read built-in command help: %w", err)
		}
		return result.Params, nil
	}
	data, err := os.ReadFile(strings.TrimSpace(path))
	if err != nil {
		return nil, err
	}
	var manifest struct {
		Profiles []struct {
			Name     string
			Commands []struct {
				Name    string
				Private bool
				Help    struct{ Variables []specVariable }
			}
		}
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	for _, profile := range manifest.Profiles {
		if profile.Name != parts[len(parts)-1] {
			continue
		}
		for _, cmd := range profile.Commands {
			if cmd.Name == name && !cmd.Private {
				return cmd.Help.Variables, nil
			}
		}
	}
	return nil, fmt.Errorf("no public command definition for %s in %s", strings.Join(command, " "), path)
}

var staticProfile = regexp.MustCompile(`^1 profile:[[:alnum:]_:.@/-]+$`)

func inspectCommand(path []string, flags ...string) ([]byte, error) {
	args := append(append([]string{}, path...), flags...)
	cmd := exec.Command("aux4", args...)
	cmd.Env = append(os.Environ(), "AUX4_NO_HOOKS=true", "AUX4_NO_DAEMON=1")
	data, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("inspect %s: %w", strings.Join(path, " "), err)
	}
	return data, nil
}
