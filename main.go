package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: aux4 config <command> [options]\n")
		fmt.Fprintf(os.Stderr, "Commands: get, set, merge, spec generate\n")
		os.Exit(1)
	}

	if os.Args[1] == "spec-generate" {
		var command []string
		if len(os.Args) != 5 || json.Unmarshal([]byte(os.Args[4]), &command) != nil || len(command) == 0 {
			fmt.Fprintln(os.Stderr, "Usage: aux4 config spec generate <command...> [--file config.spec.yaml] [--prefix path]")
			os.Exit(1)
		}
		variables, err := commandVariables(command)
		if err == nil {
			err = mergeSpec(os.Args[2], os.Args[3], variables)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	configFile := ""
	if len(os.Args) > 2 {
		configFile = os.Args[2]
	}

	if configFile == "" {
		configFile = findConfigFile()
	}

	if configFile == "" {
		fmt.Fprintf(os.Stderr, "Error: No config file found (config.json, config.yaml, or config.yml)\n")
		os.Exit(1)
	}

	aux4Config, err := loadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "get":
		getConfig(aux4Config)
	case "set":
		setConfig(aux4Config, configFile)
	case "merge":
		mergeAux4Config(aux4Config, configFile)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Fprintf(os.Stderr, "Available commands: get, set, merge, spec generate\n")
		os.Exit(1)
	}
}
