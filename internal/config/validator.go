package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateActions validates a slice of actions
// Returns error if any validation fails, collecting all errors before returning
func ValidateActions(srcRoot string, actions []Action) error {
	var errs []string

	for _, action := range actions {
		switch action.Type {
		case ActionCopy, ActionSymlink:
			// Check if source file exists
			fullPath := filepath.Join(srcRoot, action.Path)
			if _, err := os.Stat(fullPath); err != nil {
				errs = append(errs, fmt.Sprintf("line %d: source file not found: %s", action.Line, action.Path))
			}

		case ActionRun:
			// Check that command is not empty
			if action.Path == "" {
				errs = append(errs, fmt.Sprintf("line %d: run command cannot be empty", action.Line))
			}

		default:
			// Unknown action type
			errs = append(errs, fmt.Sprintf("line %d: unknown action type: %s", action.Line, action.Type))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation failed:\n  %s", strings.Join(errs, "\n  "))
	}

	return nil
}
