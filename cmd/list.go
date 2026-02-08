package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/a-tarek/ptt/internal/git"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var showAllPaths bool

var lsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List all worktrees",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if inside git repo
		if !git.IsInsideGitRepo() {
			return fmt.Errorf("not inside a git repository")
		}

		// Get all worktrees
		worktrees, err := git.ListWorktrees()
		if err != nil {
			return err
		}

		// Empty state: silent exit
		if len(worktrees) == 0 {
			return nil
		}

		// Get current worktree path
		currentPath, _ := git.CurrentWorktreeRoot()

		// Color setup (auto-detects TTY)
		green := color.New(color.FgGreen)
		yellow := color.New(color.FgYellow)
		normalColor := color.New(color.Reset)

		// Display each worktree
		for _, wt := range worktrees {
			// Marker for current worktree
			marker := " "
			if wt.Path == currentPath {
				marker = "*"
			}

			// Get directory name
			name := filepath.Base(wt.Path)

			// Check dirty/clean status
			dirty, _ := git.IsDirty(wt.Path)
			status := " "
			if dirty {
				status = "~"
			}

			// Format output with colors
			if wt.Path == currentPath {
				// Current worktree in green
				if showAllPaths {
					if dirty {
						fmt.Printf("%s ", marker)
						green.Printf("%-30s", name)
						fmt.Printf(" %-20s ", wt.Branch)
						yellow.Printf("%s", status)
						fmt.Printf(" %s\n", wt.Path)
					} else {
						fmt.Printf("%s ", marker)
						green.Printf("%-30s", name)
						fmt.Printf(" %-20s %s %s\n", wt.Branch, status, wt.Path)
					}
				} else {
					if dirty {
						fmt.Printf("%s ", marker)
						green.Printf("%-30s", name)
						fmt.Printf(" %-20s ", wt.Branch)
						yellow.Printf("%s\n", status)
					} else {
						fmt.Printf("%s ", marker)
						green.Printf("%-30s", name)
						fmt.Printf(" %-20s %s\n", wt.Branch, status)
					}
				}
			} else {
				// Other worktrees
				if showAllPaths {
					if dirty {
						normalColor.Printf("%s %-30s %-20s ", marker, name, wt.Branch)
						yellow.Printf("%s", status)
						fmt.Printf(" %s\n", wt.Path)
					} else {
						normalColor.Printf("%s %-30s %-20s %s %s\n", marker, name, wt.Branch, status, wt.Path)
					}
				} else {
					if dirty {
						normalColor.Printf("%s %-30s %-20s ", marker, name, wt.Branch)
						yellow.Printf("%s\n", status)
					} else {
						normalColor.Printf("%s %-30s %-20s %s\n", marker, name, wt.Branch, status)
					}
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
	lsCmd.Flags().BoolVarP(&showAllPaths, "all", "a", false, "show full paths")
}
