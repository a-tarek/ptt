package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveConfigPath resolves a config name/path to an actual file path
// - Empty name -> try .pttconfig/default, fall back to .wtconfig
// - Bare name (no "/") -> try .pttconfig/{name}, fall back to .wtconfig-{name}
// - Contains "/" -> treated as exact path
// Returns error if resolved path doesn't exist
func ResolveConfigPath(repoRoot string, name string) (string, error) {
	var candidates []string

	if name == "" {
		// Default config: try .pttconfig/default first, then .wtconfig
		candidates = []string{
			filepath.Join(repoRoot, ".pttconfig", "default"),
			filepath.Join(repoRoot, ".wtconfig"),
		}
	} else if strings.Contains(name, "/") {
		// Exact path (contains "/")
		candidates = []string{name}
	} else {
		// Bare name: try .pttconfig/{name} first, then .wtconfig-{name}
		candidates = []string{
			filepath.Join(repoRoot, ".pttconfig", name),
			filepath.Join(repoRoot, ".wtconfig-"+name),
		}
	}

	// Try each candidate in order
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	// None found - return error with first candidate
	return "", fmt.Errorf("config file not found: %s", candidates[0])
}
