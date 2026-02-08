package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/a-tarek/ptt/internal/config"
	"github.com/a-tarek/ptt/internal/ui"
)

// ExecuteActions runs all actions sequentially against the target worktree.
// srcRoot is the source worktree (where files are copied/symlinked from).
// targetRoot is the new worktree (where files land and commands run).
// tasks is the TaskList tracking all planned steps; offset is the index of the
// first action within that list.
// If any action fails, the entire worktree is rolled back.
func ExecuteActions(srcRoot, targetRoot string, actions []config.Action, tasks *ui.TaskList, offset int) error {
	if len(actions) == 0 {
		return nil
	}

	for i, action := range actions {
		if err := executeOne(srcRoot, targetRoot, action); err != nil {
			tasks.MarkFailed(offset + i)
			tasks.FailRemaining(offset + i + 1)
			// Rollback: remove entire worktree
			rollbackWorktree(targetRoot)
			return err
		}
		tasks.MarkDone(offset + i)
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

	case config.ActionSymlink:
		src := filepath.Join(srcRoot, action.Path)
		dest := filepath.Join(targetRoot, action.Path)
		if err := CreateSymlink(src, dest); err != nil {
			return err
		}

	case config.ActionRun:
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
	removeCmd := exec.Command("git", "worktree", "remove", "--force", targetRoot)
	if err := removeCmd.Run(); err != nil {
		// Git worktree remove failed - try direct directory removal
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
