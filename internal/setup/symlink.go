package setup

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateSymlink creates a symlink at dest pointing to src, creating parent dirs as needed
func CreateSymlink(src, dest string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	if err := os.Symlink(src, dest); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}
	return nil
}
