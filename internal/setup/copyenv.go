package setup

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/a-tarek/ptt/internal/config"
	"github.com/a-tarek/ptt/internal/envfile"
	"github.com/a-tarek/ptt/internal/git"
)

// ResolveIncrementValue scans worktrees for the max integer value of a var.
// Skips worktrees where the file or var is missing. Returns max+1.
func ResolveIncrementValue(worktrees []git.Worktree, envFilePath, varName string) (int, error) {
	max := 0
	found := false

	for _, wt := range worktrees {
		path := filepath.Join(wt.Path, envFilePath)
		ef, err := envfile.Parse(path)
		if err != nil {
			continue // skip worktrees where file doesn't exist
		}
		val, exists := ef.Get(varName)
		if !exists {
			continue // skip worktrees missing the var
		}
		n, err := strconv.Atoi(val)
		if err != nil {
			continue // skip non-integer values
		}
		if !found || n > max {
			max = n
			found = true
		}
	}

	if !found {
		return 0, fmt.Errorf("no worktree has integer value for %s in %s", varName, envFilePath)
	}

	return max + 1, nil
}

// RunStrategyCommand runs a shell command and returns trimmed stdout.
func RunStrategyCommand(workDir, command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("strategy command failed with exit code %d: %s", exitErr.ExitCode(), strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("strategy command failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ExecuteCopyEnv copies an env file then applies variable transformations.
func ExecuteCopyEnv(srcRoot, targetRoot string, cfg *config.CopyEnvConfig) error {
	src := filepath.Join(srcRoot, cfg.File)
	dest := filepath.Join(targetRoot, cfg.File)

	// Step 1: Copy the file
	if err := CopyPath(src, dest); err != nil {
		return fmt.Errorf("copyEnv: %w", err)
	}

	// Step 2: Parse the copied file
	ef, err := envfile.Parse(dest)
	if err != nil {
		return fmt.Errorf("copyEnv: failed to parse %s: %w", dest, err)
	}

	// Step 3: Apply config-defined var strategies
	for varName, rule := range cfg.Vars {
		// Skip if overridden by --set
		if _, overridden := cfg.Overrides[varName]; overridden {
			continue
		}

		value, err := resolveStrategy(targetRoot, cfg.File, varName, rule.Strategy)
		if err != nil {
			return fmt.Errorf("copyEnv: var %s: %w", varName, err)
		}
		if err := ef.Set(varName, value); err != nil {
			return fmt.Errorf("copyEnv: %w", err)
		}
	}

	// Step 4: Apply --set overrides
	for varName, override := range cfg.Overrides {
		value, err := resolveOverride(targetRoot, cfg.File, varName, override)
		if err != nil {
			return fmt.Errorf("copyEnv: var %s: %w", varName, err)
		}
		if err := ef.Set(varName, value); err != nil {
			return fmt.Errorf("copyEnv: %w", err)
		}
	}

	// Step 5: Write modified file
	if err := ef.Write(dest); err != nil {
		return fmt.Errorf("copyEnv: failed to write %s: %w", dest, err)
	}

	return nil
}

func resolveStrategy(workDir, envFilePath, varName, strategy string) (string, error) {
	if strategy == "ptt_increment" {
		worktrees, err := git.ListWorktrees()
		if err != nil {
			return "", fmt.Errorf("failed to list worktrees: %w", err)
		}
		n, err := ResolveIncrementValue(worktrees, envFilePath, varName)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(n), nil
	}

	// Shell command strategy
	return RunStrategyCommand(workDir, strategy)
}

func resolveOverride(workDir, envFilePath, varName, override string) (string, error) {
	if override == "ptt_increment" {
		return resolveStrategy(workDir, envFilePath, varName, "ptt_increment")
	}
	// Direct value
	return override, nil
}
