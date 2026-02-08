package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/a-tarek/ptt/internal/git"
	"github.com/spf13/cobra"
)

var rebaseCmd = &cobra.Command{
	Use:               "rebase <worktree>",
	Short:             "Rebase current onto worktree's branch",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: worktreeNameCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !git.IsInsideGitRepo() {
			return fmt.Errorf("not inside a git repository")
		}

		// Resolve worktree
		wt, err := git.ResolveWorktree(args[0])
		if err != nil {
			return err
		}

		// Error if worktree has no branch
		if wt.Branch == "" {
			return fmt.Errorf("worktree '%s' has no branch", args[0])
		}

		// Print status message
		fmt.Fprintf(os.Stderr, "Rebasing onto %s...\n", wt.Branch)

		// Run git rebase
		rebaseCmd := exec.Command("git", "rebase", wt.Branch)
		rebaseCmd.Stdout = os.Stdout
		rebaseCmd.Stderr = os.Stderr
		if err := rebaseCmd.Run(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(rebaseCmd)
}
