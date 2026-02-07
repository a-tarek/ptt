package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ahmedelarabyy/wt/internal/git"
	"github.com/spf13/cobra"
)

var outputPath bool

var gotoCmd = &cobra.Command{
	Use:               "goto <worktree>",
	Short:             "Switch to a worktree",
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
	rootCmd.AddCommand(gotoCmd)
	// Register --output-path as a hidden persistent flag on root
	rootCmd.PersistentFlags().BoolVar(&outputPath, "output-path", false, "output only the target path")
	rootCmd.PersistentFlags().MarkHidden("output-path")
}
