package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/a-tarek/ptt/internal/envfile"
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

		case ActionCopyEnv:
			errs = append(errs, validateCopyEnvAction(srcRoot, action)...)

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

// ValidateRemoveActions validates actions for the remove lifecycle phase.
// Only "run" actions are allowed in remove hooks.
func ValidateRemoveActions(actions []Action) error {
	var errs []string

	for i, action := range actions {
		switch action.Type {
		case ActionRun:
			if action.Path == "" {
				errs = append(errs, fmt.Sprintf("remove[%d]: run command cannot be empty", i))
			}
		case ActionCopy:
			errs = append(errs, fmt.Sprintf("remove[%d]: copy actions not allowed in remove hooks", i))
		case ActionSymlink:
			errs = append(errs, fmt.Sprintf("remove[%d]: symlink actions not allowed in remove hooks", i))
		case ActionCopyEnv:
			errs = append(errs, fmt.Sprintf("remove[%d]: copyEnv actions not allowed in remove hooks", i))
		default:
			errs = append(errs, fmt.Sprintf("remove[%d]: unknown action type: %s", i, action.Type))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation failed:\n  %s", strings.Join(errs, "\n  "))
	}

	return nil
}

// validateCopyEnvAction validates a copyEnv action, collecting all errors.
func validateCopyEnvAction(srcRoot string, action Action) []string {
	var errs []string
	cfg := action.CopyEnv
	if cfg == nil {
		errs = append(errs, fmt.Sprintf("line %d: copyEnv action missing config", action.Line))
		return errs
	}

	// Check source file exists
	fullPath := filepath.Join(srcRoot, cfg.File)
	if _, err := os.Stat(fullPath); err != nil {
		errs = append(errs, fmt.Sprintf("line %d: source file not found: %s", action.Line, cfg.File))
		return errs // Can't validate vars without the file
	}

	// Parse env file to validate vars
	ef, err := envfile.Parse(fullPath)
	if err != nil {
		errs = append(errs, fmt.Sprintf("line %d: failed to parse env file %s: %v", action.Line, cfg.File, err))
		return errs
	}

	// Validate config-defined vars
	for varName, rule := range cfg.Vars {
		val, exists := ef.Get(varName)
		if !exists {
			errs = append(errs, fmt.Sprintf("line %d: var %s not found in %s", action.Line, varName, cfg.File))
			continue
		}
		if strings.HasPrefix(rule.Strategy, "ptt_") {
			if rule.Strategy == "ptt_increment" {
				if _, err := strconv.Atoi(val); err != nil {
					errs = append(errs, fmt.Sprintf("line %d: var %s must be integer for ptt_increment, got %q", action.Line, varName, val))
				}
			}
		} else if rule.Strategy == "" {
			errs = append(errs, fmt.Sprintf("line %d: var %s has empty strategy", action.Line, varName))
		}
	}

	// Validate --set overrides
	for varName, value := range cfg.Overrides {
		_, exists := ef.Get(varName)
		if !exists {
			errs = append(errs, fmt.Sprintf("line %d: --set var %s not found in %s", action.Line, varName, cfg.File))
			continue
		}
		if value == "ptt_increment" {
			val, _ := ef.Get(varName)
			if _, err := strconv.Atoi(val); err != nil {
				errs = append(errs, fmt.Sprintf("line %d: var %s must be integer for ptt_increment, got %q", action.Line, varName, val))
			}
		}
	}

	return errs
}
