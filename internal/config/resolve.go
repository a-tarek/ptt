package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveConfigPath resolves a config name/path to an actual file path
// - Empty name -> {repoRoot}/.wtconfig
// - Bare name (no "/") -> {repoRoot}/.wtconfig-{name}
// - Contains "/" -> treated as exact path
// Returns error if resolved path doesn't exist
func ResolveConfigPath(repoRoot string, name string) (string, error) {
	var configPath string

	if name == "" {
		// Default config
		configPath = filepath.Join(repoRoot, ".wtconfig")
	} else if strings.Contains(name, "/") {
		// Exact path (contains "/")
		configPath = name
	} else {
		// Bare name -> .wtconfig-{name}
		configPath = filepath.Join(repoRoot, ".wtconfig-"+name)
	}

	// Check if file exists
	if _, err := os.Stat(configPath); err != nil {
		return "", fmt.Errorf("config file not found: %s", configPath)
	}

	return configPath, nil
}
