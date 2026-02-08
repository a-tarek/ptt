package config

import (
	"fmt"
	"strings"
)

// BuildActionsFromFlags creates an ordered Action slice from flag inputs.
// Uses osArgs to preserve the original flag order from the command line.
func BuildActionsFromFlags(copyPaths, symlinkPaths, runCommands, osArgs []string) []Action {
	// Build lookup maps: flag value -> Action
	copyMap := make(map[string]bool)
	for _, p := range copyPaths {
		copyMap[p] = true
	}
	symlinkMap := make(map[string]bool)
	for _, p := range symlinkPaths {
		symlinkMap[p] = true
	}
	runMap := make(map[string]bool)
	for _, c := range runCommands {
		runMap[c] = true
	}

	// Walk osArgs to reconstruct original order
	var actions []Action
	for i := 0; i < len(osArgs); i++ {
		arg := osArgs[i]
		var flagType, value string

		if arg == "--copy" || arg == "--symlink" || arg == "--run" {
			flagType = strings.TrimPrefix(arg, "--")
			if i+1 < len(osArgs) {
				i++
				value = osArgs[i]
			}
		} else if strings.HasPrefix(arg, "--copy=") {
			flagType = "copy"
			value = strings.TrimPrefix(arg, "--copy=")
		} else if strings.HasPrefix(arg, "--symlink=") {
			flagType = "symlink"
			value = strings.TrimPrefix(arg, "--symlink=")
		} else if strings.HasPrefix(arg, "--run=") {
			flagType = "run"
			value = strings.TrimPrefix(arg, "--run=")
		} else {
			continue
		}

		switch flagType {
		case "copy":
			if copyMap[value] {
				actions = append(actions, Action{Type: ActionCopy, Path: value})
				delete(copyMap, value)
			}
		case "symlink":
			if symlinkMap[value] {
				actions = append(actions, Action{Type: ActionSymlink, Path: value})
				delete(symlinkMap, value)
			}
		case "run":
			if runMap[value] {
				actions = append(actions, Action{Type: ActionRun, Path: value})
				delete(runMap, value)
			}
		}
	}

	// Append any remaining values not found in osArgs (safety fallback)
	for _, p := range copyPaths {
		if copyMap[p] {
			actions = append(actions, Action{Type: ActionCopy, Path: p})
		}
	}
	for _, p := range symlinkPaths {
		if symlinkMap[p] {
			actions = append(actions, Action{Type: ActionSymlink, Path: p})
		}
	}
	for _, c := range runCommands {
		if runMap[c] {
			actions = append(actions, Action{Type: ActionRun, Path: c})
		}
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
