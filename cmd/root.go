package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wt",
	Short: "Git worktree manager",
	Long: `wt — Git Worktree Manager

A fast, intuitive git worktree manager.

Commands:
  new [flags] <name> [branch]                  Create a new worktree
  goto <worktree>                              cd into a worktree
  home                                         cd into the main worktree
  init                                         Create .wtconfig template
  eject [flags] [name]                         Eject current branch into its own worktree
  list                                         List all worktrees
  merge <worktree>                             Merge worktree's branch into current
  rebase <worktree>                            Rebase current onto worktree's branch
  delete <worktree>                            Remove a worktree (keeps branch)`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return err
	}
	return nil
}
