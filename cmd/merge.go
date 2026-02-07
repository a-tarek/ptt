package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/ahmedelarabyy/wt/internal/git"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:               "merge <worktree>",
	Short:             "Merge a worktree's branch into current",
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
		fmt.Fprintf(os.Stderr, "Merging %s into current branch...\n", wt.Branch)

		// Run git merge
		mergeCmd := exec.Command("git", "merge", wt.Branch)
		mergeCmd.Stdout = os.Stdout
		mergeCmd.Stderr = os.Stderr
		if err := mergeCmd.Run(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}
