package cmd

import (
	"fmt"
	"os"

	"github.com/a-tarek/ptt/internal/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ptt",
	Short: "Git worktree manager",
	Long: `ptt — Git Worktree Manager

A fast, intuitive git worktree manager.

Commands:
  mk [flags] <name>                            Create a new worktree
  cd [worktree]                                Navigate to a worktree (or home)
  init                                         Initialize repository for ptt
  eject [flags] [name]                         Eject current branch
  ls                                           List all worktrees
  merge <worktree>                             Merge worktree's branch
  rebase <worktree>                            Rebase onto worktree's branch
  rm <worktree>                                Remove a worktree`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		// Extract subcommand name from args
		cmd := ""
		if len(os.Args) > 1 {
			// os.Args[1] is the subcommand (e.g., "cd", "delete")
			// Ignore if it starts with "-" (flag) or is "help"
			arg := os.Args[1]
			if arg != "" && arg[0] != '-' && arg != "help" {
				cmd = arg
			}
		}

		fmt.Fprintf(os.Stderr, "%s\n", ui.FormatError(err, cmd))
		return err
	}
	return nil
}
