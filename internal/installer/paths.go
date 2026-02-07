package installer

import (
	"fmt"
	"os"
	"path/filepath"
)

// RCFilePath returns the default rc file path for a given shell type.
// Returns an error for unsupported shells.
func RCFilePath(shellType string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	switch shellType {
	case "zsh":
		return filepath.Join(homeDir, ".zshrc"), nil
	case "bash":
		// On macOS, .bash_profile is commonly used instead of .bashrc
		// If .bash_profile exists but .bashrc does not, prefer .bash_profile
		bashrc := filepath.Join(homeDir, ".bashrc")
		bashProfile := filepath.Join(homeDir, ".bash_profile")

		// Check if .bashrc exists
		if _, err := os.Stat(bashrc); err == nil {
			return bashrc, nil
		}

		// Check if .bash_profile exists
		if _, err := os.Stat(bashProfile); err == nil {
			return bashProfile, nil
		}

		// Neither exists, default to .bashrc (will be created)
		return bashrc, nil
	case "fish":
		return filepath.Join(homeDir, ".config", "fish", "config.fish"), nil
	default:
		return "", fmt.Errorf("unsupported shell type: %s (supported: bash, zsh, fish)", shellType)
	}
}
