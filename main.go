package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: aux4 config <command> [options]\n")
		fmt.Fprintf(os.Stderr, "Commands: get, set, merge\n")
		os.Exit(1)
	}

	// Find config file
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

	// Load config
	aux4Config, err := loadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Handle commands
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
		fmt.Fprintf(os.Stderr, "Available commands: get, set, merge\n")
		os.Exit(1)
	}
}