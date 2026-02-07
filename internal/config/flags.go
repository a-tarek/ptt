package config

import "fmt"

// BuildActionsFromFlags creates an ordered Action slice from flag inputs
// Order: all copies, then all symlinks, then all runs
// (Cobra cannot preserve interleaved order across different flag types)
func BuildActionsFromFlags(copyPaths, symlinkPaths, runCommands []string) []Action {
	var actions []Action

	// Add copy actions
	for _, path := range copyPaths {
		actions = append(actions, Action{
			Type: ActionCopy,
			Path: path,
			Line: 0, // from flags, not file
		})
	}

	// Add symlink actions
	for _, path := range symlinkPaths {
		actions = append(actions, Action{
			Type: ActionSymlink,
			Path: path,
			Line: 0, // from flags, not file
		})
	}

	// Add run actions
	for _, cmd := range runCommands {
		actions = append(actions, Action{
			Type: ActionRun,
			Path: cmd,
			Line: 0, // from flags, not file
		})
	}

	return actions
}

// CheckDuplicatePaths checks for duplicate paths across copy and symlink flags
// Run commands are not checked (they're commands, not file paths)
func CheckDuplicatePaths(copyPaths, symlinkPaths []string) error {
	seen := make(map[string]string) // path -> action type

	// Check copy paths
	for _, path := range copyPaths {
		if prevType, exists := seen[path]; exists {
			return fmt.Errorf("duplicate path '%s' in --%s and --copy", path, prevType)
		}
		seen[path] = "copy"
	}

	// Check symlink paths
	for _, path := range symlinkPaths {
		if prevType, exists := seen[path]; exists {
			return fmt.Errorf("duplicate path '%s' in --%s and --symlink", path, prevType)
		}
		seen[path] = "symlink"
	}

	return nil
}
