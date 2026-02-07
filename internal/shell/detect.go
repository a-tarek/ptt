package shell

import (
	"fmt"
	"os"
	"path/filepath"
)

// DetectShell returns the shell type based on the SHELL environment variable.
// Returns "bash", "zsh", or "fish" for supported shells, or an error otherwise.
func DetectShell() (string, error) {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return "", fmt.Errorf("SHELL environment variable not set")
	}

	shellName := filepath.Base(shellPath)

	switch shellName {
	case "bash", "zsh", "fish":
		return shellName, nil
	default:
		return "", fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shellName)
	}
}
