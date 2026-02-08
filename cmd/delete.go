package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/a-tarek/ptt/internal/git"
	"github.com/spf13/cobra"
)

var (
	forceDelete  bool
	deleteBranch bool
)

var deleteCmd = &cobra.Command{
	Use:               "delete <worktree>",
	Short:             "Remove a worktree",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: worktreeNameCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Check if inside git repo
		if !git.IsInsideGitRepo() {
			return fmt.Errorf("not inside a git repository")
		}

		// Resolve worktree by name
		wt, err := git.ResolveWorktree(name)
		if err != nil {
			return err
		}

		// Check if trying to delete current worktree
		currentPath, err := git.CurrentWorktreeRoot()
		if err != nil {
			return err
		}
		if wt.Path == currentPath {
			return fmt.Errorf("can't delete current worktree")
		}

		// Check if worktree is dirty
		dirty, err := git.IsDirty(wt.Path)
		if err != nil {
			return err
		}

		// If dirty and not --force, prompt for confirmation
		if dirty && !forceDelete {
			basename := filepath.Base(wt.Path)
			fmt.Fprintf(os.Stderr, "Worktree '%s' has uncommitted changes. Delete? [y/N] ", basename)

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return err
			}

			response = strings.TrimSpace(response)
			if response != "y" && response != "Y" {
				return fmt.Errorf("cancelled")
			}
		}

		// Remove worktree
		var removeCmd *exec.Cmd
		if dirty {
			// Force remove if dirty (user confirmed or --force was passed)
			removeCmd = exec.Command("git", "worktree", "remove", "--force", wt.Path)
		} else {
			removeCmd = exec.Command("git", "worktree", "remove", wt.Path)
		}

		output, err := removeCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to remove worktree: %s", string(output))
		}

		// If --branch flag is set, also delete the branch
		if deleteBranch && wt.Branch != "" {
			// Try safe delete first
			branchCmd := exec.Command("git", "branch", "-d", wt.Branch)
			err := branchCmd.Run()

			if err != nil {
				// Branch is not fully merged
				if forceDelete {
					// Force delete if --force was passed
					branchCmd = exec.Command("git", "branch", "-D", wt.Branch)
					if err := branchCmd.Run(); err != nil {
						// Non-fatal warning
						fmt.Fprintf(os.Stderr, "warning: failed to delete branch '%s'\n", wt.Branch)
					}
				} else {
					// Warn but don't fail
					fmt.Fprintf(os.Stderr, "warning: branch '%s' not fully merged, use --force to delete\n", wt.Branch)
				}
			}
		}

		// Silent success
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "skip confirmation for dirty worktrees")
	deleteCmd.Flags().BoolVarP(&deleteBranch, "branch", "b", false, "also delete the branch after removing worktree")
}
