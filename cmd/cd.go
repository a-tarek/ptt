package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/a-tarek/ptt/internal/git"
	"github.com/spf13/cobra"
)

var outputPath bool

var cdCmd = &cobra.Command{
	Use:               "cd [worktree]",
	Aliases:           []string{},
	Short:             "Navigate to a worktree (or home if no args)",
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: worktreeNameCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !git.IsInsideGitRepo() {
			return fmt.Errorf("not inside a git repository")
		}

		// Handle home case (zero args)
		if len(args) == 0 {
			// Get home path
			homePath, err := git.GetHomePath()
			if err != nil {
				return err
			}

			// Check if already home
			currentPath, err := git.CurrentWorktreeRoot()
			if err != nil {
				return err
			}

			if homePath == currentPath {
				fmt.Fprintf(os.Stderr, "Already home\n")
				return nil
			}

			// Get basename for display
			basename := filepath.Base(homePath)

			// Check dirty status
			dirty, err := git.IsDirty(homePath)
			if err != nil {
				return err
			}

			dirtyStr := "clean"
			if dirty {
				dirtyStr = "dirty"
			}

			// Get branch name from worktree list (more reliable than CurrentBranch for home path)
			worktrees, err := git.ListWorktrees()
			if err != nil {
				return err
			}

			branchStr := "(bare)"
			for _, wt := range worktrees {
				if wt.Path == homePath {
					if wt.Branch != "" {
						branchStr = wt.Branch
					}
					break
				}
			}

			// Output path for shell wrapper
			if outputPath {
				fmt.Println(homePath)
			}

			// Always print confirmation to stderr
			fmt.Fprintf(os.Stderr, "Switched to %s (branch: %s, %s)\n", basename, branchStr, dirtyStr)

			// If not using output-path mode, also print confirmation to stdout
			if !outputPath {
				fmt.Printf("Switched to %s (branch: %s, %s)\n", basename, branchStr, dirtyStr)
			}

			return nil
		}

		// Handle goto case (one arg)
		// Resolve worktree
		wt, err := git.ResolveWorktree(args[0])
		if err != nil {
			return err
		}

		// Check if already in target worktree
		currentPath, err := git.CurrentWorktreeRoot()
		if err != nil {
			return err
		}

		basename := filepath.Base(wt.Path)
		if wt.Path == currentPath {
			fmt.Fprintf(os.Stderr, "Already in %s\n", basename)
			return nil
		}

		// Check dirty status
		dirty, err := git.IsDirty(wt.Path)
		if err != nil {
			return err
		}

		dirtyStr := "clean"
		if dirty {
			dirtyStr = "dirty"
		}

		// Output path for shell wrapper
		if outputPath {
			fmt.Println(wt.Path)
		}

		// Always print confirmation to stderr
		fmt.Fprintf(os.Stderr, "Switched to %s (branch: %s, %s)\n", basename, wt.Branch, dirtyStr)

		// If not using output-path mode, also print confirmation to stdout
		if !outputPath {
			fmt.Printf("Switched to %s (branch: %s, %s)\n", basename, wt.Branch, dirtyStr)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cdCmd)
	// Register --output-path as a hidden persistent flag on root
	rootCmd.PersistentFlags().BoolVar(&outputPath, "output-path", false, "output only the target path")
	rootCmd.PersistentFlags().MarkHidden("output-path")
}
