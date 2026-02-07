package setup

import (
	"fmt"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
)

// CopyPath copies a file or directory from src to dest, creating parent dirs as needed
func CopyPath(src, dest string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	if err := cp.Copy(src, dest); err != nil {
		return fmt.Errorf("failed to copy %s: %w", src, err)
	}
	return nil
}
