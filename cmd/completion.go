package cmd

import (
	"path/filepath"

	"github.com/a-tarek/ptt/internal/git"
	"github.com/spf13/cobra"
)

// worktreeNameCompletion provides dynamic worktree name completions for commands
// that accept a worktree name as their first positional argument.
func worktreeNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Don't complete if we already have the required positional arg
	if len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Query git worktree list (live, no caching)
	worktrees, err := git.ListWorktrees()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Extract worktree basenames for completion
	// Show names only (no branch descriptions) per user decision
	names := make([]string, 0, len(worktrees))
	for _, wt := range worktrees {
		name := filepath.Base(wt.Path)
		if name != "" && name != "." {
			names = append(names, name)
		}
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

// branchNameCompletion provides dynamic branch name completions for --branch flag.
// It returns all branches minus those already checked out in existing worktrees.
func branchNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	branches, err := git.ListBranches()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Get branches already checked out in worktrees
	worktrees, err := git.ListWorktrees()
	if err != nil {
		return branches, cobra.ShellCompDirectiveNoFileComp
	}

	checkedOut := make(map[string]bool, len(worktrees))
	for _, wt := range worktrees {
		if wt.Branch != "" {
			checkedOut[wt.Branch] = true
		}
	}

	available := make([]string, 0, len(branches))
	for _, b := range branches {
		if !checkedOut[b] {
			available = append(available, b)
		}
	}

	return available, cobra.ShellCompDirectiveNoFileComp
}
