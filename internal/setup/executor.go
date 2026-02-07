package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ahmedelarabyy/wt/internal/config"
)

// ExecuteActions runs all actions sequentially against the target worktree.
// srcRoot is the source worktree (where files are copied/symlinked from).
// targetRoot is the new worktree (where files land and commands run).
// If any action fails, the entire worktree is rolled back.
func ExecuteActions(srcRoot, targetRoot string, actions []config.Action) error {
	if len(actions) == 0 {
		return nil
	}

	for _, action := range actions {
		if err := executeOne(srcRoot, targetRoot, action); err != nil {
			// Rollback: remove entire worktree
			rollbackWorktree(targetRoot)
			return err
		}
	}

	return nil
}

// executeOne executes a single action
func executeOne(srcRoot, targetRoot string, action config.Action) error {
	switch action.Type {
	case config.ActionCopy:
		src := filepath.Join(srcRoot, action.Path)
		dest := filepath.Join(targetRoot, action.Path)
		if err := CopyPath(src, dest); err != nil {
			return err
		}
		fmt.Printf("Copied %s\n", action.Path)

	case config.ActionSymlink:
		src := filepath.Join(srcRoot, action.Path)
		dest := filepath.Join(targetRoot, action.Path)
		if err := CreateSymlink(src, dest); err != nil {
			return err
		}
		fmt.Printf("Symlinked %s\n", action.Path)

	case config.ActionRun:
		fmt.Printf("Running %s...\n", action.Path)
		if err := RunCommand(targetRoot, action.Path); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}

	return nil
}

// rollbackWorktree removes the worktree directory
func rollbackWorktree(targetRoot string) {
	fmt.Fprintf(os.Stderr, "Error occurred, rolling back worktree...\n")

	// Try to remove the worktree using git
	// Note: This assumes we're being called from the context of the wt command
	// which will handle setting the correct working directory for git commands
	removeCmd := exec.Command("git", "worktree", "remove", "--force", targetRoot)
	if err := removeCmd.Run(); err != nil {
		// Git worktree remove failed - try direct directory removal
		// This is a fallback for edge cases
		if err := os.RemoveAll(targetRoot); err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: rollback failed: %v\n", err)
			fmt.Fprintf(os.Stderr, "Manual cleanup: git worktree remove --force %s\n", targetRoot)
			fmt.Fprintf(os.Stderr, "Or: rm -rf %s\n", targetRoot)
		}
		return
	}

	// Clean up directory if git left it
	if _, err := os.Stat(targetRoot); err == nil {
		if err := os.RemoveAll(targetRoot); err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: directory cleanup failed: %v\n", err)
			fmt.Fprintf(os.Stderr, "Manual cleanup: rm -rf %s\n", targetRoot)
		}
	}
}
