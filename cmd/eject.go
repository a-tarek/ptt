package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/a-tarek/ptt/internal/config"
	"github.com/a-tarek/ptt/internal/git"
	"github.com/a-tarek/ptt/internal/setup"
	"github.com/a-tarek/ptt/internal/ui"
	"github.com/spf13/cobra"
)

var (
	ejectConfigName  string
	ejectSkipConfig  bool
	ejectCopyFlags   []string
	ejectSymlinkFlags []string
	ejectRunFlags    []string
)

var ejectCmd = &cobra.Command{
	Use:   "eject [name]",
	Short: "Eject current branch into its own worktree",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Check git repo
		if !git.IsInsideGitRepo() {
			return fmt.Errorf("not inside a git repository")
		}

		// 2. Get current state
		currentBranch, err := git.CurrentBranch()
		if err != nil {
			return fmt.Errorf("detached HEAD -- nothing to eject")
		}

		srcRoot, err := git.CurrentWorktreeRoot()
		if err != nil {
			return err
		}

		homePath, err := git.GetHomePath()
		if err != nil {
			return err
		}

		configRoot, err := git.ConfigRoot()
		if err != nil {
			return err
		}

		// 3. Determine fallback branch
		var fallbackBranch string
		if srcRoot == homePath {
			// We're in the home worktree - fallback to main or master
			cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/main")
			if cmd.Run() == nil {
				fallbackBranch = "main"
			} else {
				cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/master")
				if cmd.Run() == nil {
					fallbackBranch = "master"
				} else {
					return fmt.Errorf("neither 'main' nor 'master' branch exists")
				}
			}
		} else {
			// We're in a non-home worktree - fallback to directory suffix
			dirName := filepath.Base(srcRoot)
			repoName := filepath.Base(homePath)
			suffix := strings.TrimPrefix(dirName, repoName+"-")
			fallbackBranch = suffix

			// Verify fallback branch exists
			cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+fallbackBranch)
			if cmd.Run() != nil {
				return fmt.Errorf("fallback branch '%s' does not exist", fallbackBranch)
			}
		}

		// Error if already on fallback branch
		if currentBranch == fallbackBranch {
			return fmt.Errorf("already on '%s' -- nothing to eject", fallbackBranch)
		}

		// 4. Determine worktree name and path
		var name string
		if len(args) > 0 {
			name = args[0]
		} else {
			name = strings.ReplaceAll(currentBranch, "/", "-")
		}

		targetPath, err := git.WorktreePath(homePath, name)
		if err != nil {
			return err
		}

		// 5. Print start message
		targetBasename := filepath.Base(targetPath)
		fmt.Fprintf(os.Stderr, "Ejecting branch '%s' into %s\n", currentBranch, targetBasename)

		// 6. Stash uncommitted changes (including untracked)
		stashBeforeCmd := exec.Command("git", "stash", "list")
		stashBeforeOutput, err := stashBeforeCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to list stashes: %w", err)
		}
		stashCountBefore := len(strings.Split(strings.TrimSpace(string(stashBeforeOutput)), "\n"))
		if strings.TrimSpace(string(stashBeforeOutput)) == "" {
			stashCountBefore = 0
		}

		stashCmd := exec.Command("git", "stash", "push", "-u", "-m", "wt-eject: "+currentBranch)
		if err := stashCmd.Run(); err != nil {
			return fmt.Errorf("failed to stash changes: %w", err)
		}

		stashAfterCmd := exec.Command("git", "stash", "list")
		stashAfterOutput, err := stashAfterCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to list stashes after stash: %w", err)
		}
		stashCountAfter := len(strings.Split(strings.TrimSpace(string(stashAfterOutput)), "\n"))
		if strings.TrimSpace(string(stashAfterOutput)) == "" {
			stashCountAfter = 0
		}

		didStash := stashCountAfter > stashCountBefore
		if didStash {
			fmt.Fprintf(os.Stderr, "Stashed uncommitted changes\n")
		}

		// 7. Switch to fallback branch
		checkoutCmd := exec.Command("git", "checkout", fallbackBranch)
		checkoutCmd.Stderr = os.Stderr
		if err := checkoutCmd.Run(); err != nil {
			// Rollback: pop stash if we stashed
			if didStash {
				popCmd := exec.Command("git", "stash", "pop")
				popCmd.Run() // ignore error
			}
			return fmt.Errorf("failed to switch to '%s'", fallbackBranch)
		}
		fmt.Fprintf(os.Stderr, "Switched to '%s'\n", fallbackBranch)

		// 8. Create new worktree
		addCmd := exec.Command("git", "worktree", "add", targetPath, currentBranch)
		addCmd.Stderr = os.Stderr
		if err := addCmd.Run(); err != nil {
			// Rollback: checkout back to original branch, pop stash if we stashed
			checkoutBackCmd := exec.Command("git", "checkout", currentBranch)
			checkoutBackCmd.Run() // ignore error
			if didStash {
				popCmd := exec.Command("git", "stash", "pop")
				popCmd.Run() // ignore error
			}
			return fmt.Errorf("failed to create worktree at %s", targetPath)
		}
		fmt.Fprintf(os.Stderr, "Created worktree: %s\n", targetBasename)

		// 9. Pop stash in new worktree
		if didStash {
			popCmd := exec.Command("git", "-C", targetPath, "stash", "pop")
			popOutput, popErr := popCmd.CombinedOutput()
			if popErr != nil {
				// Stash pop failed (likely merge conflicts) - this is non-fatal
				fmt.Fprintf(os.Stderr, "WARNING: Stash restored with conflicts -- resolve before committing\n")
			} else {
				fmt.Fprintf(os.Stderr, "Restored uncommitted changes in new worktree\n")
			}
			_ = popOutput // captured for potential debugging
		}

		// 10. Apply config (same logic as wt new)
		hasEjectInlineFlags := len(ejectCopyFlags) > 0 || len(ejectSymlinkFlags) > 0 || len(ejectRunFlags) > 0
		if !ejectSkipConfig || hasEjectInlineFlags {
			var fileActions []config.Action
			var configName string

			// File-based config (unless --skip-config or inline flags override)
			if !ejectSkipConfig && !hasEjectInlineFlags {
				configPath, err := config.ResolveConfigPath(configRoot, ejectConfigName)
				if err == nil {
					// Config file exists
					fileActions, err = config.ParseFile(configPath)
					if err != nil {
						// Config parse failed - rollback
						checkoutBackCmd := exec.Command("git", "checkout", currentBranch)
						checkoutBackCmd.Run()
						if didStash {
							popCmd := exec.Command("git", "stash", "pop")
							popCmd.Run()
						}
						removeCmd := exec.Command("git", "worktree", "remove", "--force", targetPath)
						removeCmd.Run()
						os.RemoveAll(targetPath)
						return err
					}

					// Validate file actions
					if err := config.ValidateActions(srcRoot, fileActions); err != nil {
						checkoutBackCmd := exec.Command("git", "checkout", currentBranch)
						checkoutBackCmd.Run()
						if didStash {
							popCmd := exec.Command("git", "stash", "pop")
							popCmd.Run()
						}
						removeCmd := exec.Command("git", "worktree", "remove", "--force", targetPath)
						removeCmd.Run()
						os.RemoveAll(targetPath)
						return err
					}

					if ejectConfigName == "" {
						configName = ".pttconfig/default"
					} else {
						configName = ".pttconfig/" + ejectConfigName
					}
				}
			} else if !ejectSkipConfig && hasEjectInlineFlags && ejectConfigName != "" {
				// --config with inline flags: load named config, then append inline flags
				configPath, cfgErr := config.ResolveConfigPath(configRoot, ejectConfigName)
				if cfgErr == nil {
					fileActions, err = config.ParseFile(configPath)
					if err != nil {
						checkoutBackCmd := exec.Command("git", "checkout", currentBranch)
						checkoutBackCmd.Run()
						if didStash {
							popCmd := exec.Command("git", "stash", "pop")
							popCmd.Run()
						}
						removeCmd := exec.Command("git", "worktree", "remove", "--force", targetPath)
						removeCmd.Run()
						os.RemoveAll(targetPath)
						return err
					}

					if err := config.ValidateActions(srcRoot, fileActions); err != nil {
						checkoutBackCmd := exec.Command("git", "checkout", currentBranch)
						checkoutBackCmd.Run()
						if didStash {
							popCmd := exec.Command("git", "stash", "pop")
							popCmd.Run()
						}
						removeCmd := exec.Command("git", "worktree", "remove", "--force", targetPath)
						removeCmd.Run()
						os.RemoveAll(targetPath)
						return err
					}

					configName = ".pttconfig/" + ejectConfigName
				}
			}

			// Flag-based actions
			flagActions := config.BuildActionsFromFlags(ejectCopyFlags, ejectSymlinkFlags, ejectRunFlags, os.Args)

			// Check for duplicate paths
			if err := config.CheckDuplicatePaths(ejectCopyFlags, ejectSymlinkFlags); err != nil {
				checkoutBackCmd := exec.Command("git", "checkout", currentBranch)
				checkoutBackCmd.Run()
				if didStash {
					popCmd := exec.Command("git", "stash", "pop")
					popCmd.Run()
				}
				removeCmd := exec.Command("git", "worktree", "remove", "--force", targetPath)
				removeCmd.Run()
				os.RemoveAll(targetPath)
				return err
			}

			// Combine actions (file first, then flags)
			allActions := append(fileActions, flagActions...)

			// Execute if we have any actions
			if len(allActions) > 0 {
				ejectTasks := ui.NewTaskList()
				for _, a := range allActions {
					ejectTasks.Add(a.Type+":", a.Path)
				}
				if err := setup.ExecuteActions(srcRoot, targetPath, allActions, ejectTasks, 0); err != nil {
					// ExecuteActions already removed the worktree directory via rollback
					// We also need to undo the branch switch and stash
					checkoutBackCmd := exec.Command("git", "checkout", currentBranch)
					checkoutBackCmd.Run()
					if didStash {
						popCmd := exec.Command("git", "stash", "pop")
						popCmd.Run()
					}
					return err
				}

				// Print config applied message
				actionCount := len(allActions)
				if configName != "" && len(flagActions) > 0 {
					fmt.Fprintf(os.Stderr, "Applied %s + flags (%d actions)\n", configName, actionCount)
				} else if configName != "" {
					fmt.Fprintf(os.Stderr, "Applied %s (%d actions)\n", configName, actionCount)
				} else {
					fmt.Fprintf(os.Stderr, "Applied flags (%d actions)\n", actionCount)
				}
			}
		}

		// 11. Final confirmation
		dirty, err := git.IsDirty(targetPath)
		if err != nil {
			dirty = false // ignore error, just show as clean
		}

		dirtyStr := "clean"
		if dirty {
			dirtyStr = "dirty"
		}

		// 12. Output path
		if outputPath {
			fmt.Println(targetPath)
		}

		// Always print confirmation to stderr
		fmt.Fprintf(os.Stderr, "Switched to %s (branch: %s, %s)\n", targetBasename, currentBranch, dirtyStr)

		// If not using output-path mode, also print to stdout
		if !outputPath {
			fmt.Printf("Switched to %s (branch: %s, %s)\n", targetBasename, currentBranch, dirtyStr)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(ejectCmd)
	ejectCmd.Flags().StringVar(&ejectConfigName, "config", "", "use named config (.pttconfig/<name>)")
	ejectCmd.Flags().BoolVar(&ejectSkipConfig, "skip-config", false, "skip file-based config")
	ejectCmd.Flags().StringSliceVar(&ejectCopyFlags, "copy", nil, "inline copy override")
	ejectCmd.Flags().StringSliceVar(&ejectSymlinkFlags, "symlink", nil, "inline symlink override")
	ejectCmd.Flags().StringSliceVar(&ejectRunFlags, "run", []string{}, "inline run commands (repeatable)")
}
