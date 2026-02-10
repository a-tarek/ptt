package setup

import (
	"fmt"

	"github.com/a-tarek/ptt/internal/config"
	"github.com/a-tarek/ptt/internal/ui"
)

// RunPreRemoveHooks executes remove lifecycle hooks sequentially.
// Stops on first failure and returns the error.
func RunPreRemoveHooks(worktreePath string, actions []config.Action, tasks *ui.TaskList, offset int) error {
	for i, action := range actions {
		if action.Type != config.ActionRun {
			return fmt.Errorf("unexpected action type in remove hooks: %s", action.Type)
		}

		if err := RunCommand(worktreePath, action.Path); err != nil {
			if tasks != nil {
				tasks.MarkFailed(offset + i)
				tasks.FailRemaining(offset + i + 1)
			}
			return fmt.Errorf("pre-remove hook failed: %w", err)
		}

		if tasks != nil {
			tasks.MarkDone(offset + i)
		}
	}

	return nil
}
