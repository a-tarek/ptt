package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ahmedelarabyy/wt/internal/config"
	"github.com/ahmedelarabyy/wt/internal/git"
	"github.com/ahmedelarabyy/wt/internal/setup"
	"github.com/spf13/cobra"
)

var (
	configFlag     string
	skipConfig     bool
	copyFlags      []string
	symlinkFlags   []string
	runFlags       []string
)

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new worktree",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Check git repo
		if !git.IsInsideGitRepo() {
			return fmt.Errorf("not inside a git repository")
		}

		// 2. Get repo root and current worktree root
		homePath, err := git.GetHomePath()
		if err != nil {
			return err
		}

		currentWorktreeRoot, err := git.CurrentWorktreeRoot()
		if err != nil {
			return err
		}

		// 3. Compute target path
		name := args[0]
		targetPath, err := git.WorktreePath(homePath, name)
		if err != nil {
			return err
		}

		// 4. Determine branch name
		branchName := name

		// 5. Create worktree
		// Try creating new branch first
		createCmd := exec.Command("git", "worktree", "add", targetPath, "-b", branchName)
		output, err := createCmd.CombinedOutput()
		if err != nil {
			// Branch might already exist, try without -b
			createCmd = exec.Command("git", "worktree", "add", targetPath, branchName)
			output, err = createCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to create worktree: %s", strings.TrimSpace(string(output)))
			}
		}

		// 6. Print creation message immediately
		basename := filepath.Base(targetPath)
		fmt.Fprintf(os.Stderr, "Created worktree %s (branch: %s)\n", basename, branchName)

		// 7. Handle config
		var appliedActions int
		var configFileName string
		hasInlineFlags := len(copyFlags) > 0 || len(symlinkFlags) > 0 || len(runFlags) > 0

		if !skipConfig || hasInlineFlags {
			var allActions []config.Action

			// Load file-based config (unless --skip-config or inline flags override)
			if !skipConfig && !hasInlineFlags {
				var configPath string
				if configFlag != "" {
					configPath, err = config.ResolveConfigPath(homePath, configFlag)
				} else {
					configPath, err = config.ResolveConfigPath(homePath, "")
				}

				// If config file found, parse and validate
				if err == nil {
					configFileName = filepath.Base(configPath)
					actions, parseErr := config.ParseFile(configPath)
					if parseErr != nil {
						return parseErr
					}

					validateErr := config.ValidateActions(currentWorktreeRoot, actions)
					if validateErr != nil {
						return validateErr
					}

					allActions = append(allActions, actions...)
				}
				// Silently skip if config file not found
			} else if !skipConfig && hasInlineFlags && configFlag != "" {
				// --config with inline flags: load named config, then append inline flags
				configPath, cfgErr := config.ResolveConfigPath(homePath, configFlag)
				if cfgErr == nil {
					configFileName = filepath.Base(configPath)
					actions, parseErr := config.ParseFile(configPath)
					if parseErr != nil {
						return parseErr
					}

					validateErr := config.ValidateActions(currentWorktreeRoot, actions)
					if validateErr != nil {
						return validateErr
					}

					allActions = append(allActions, actions...)
				}
			}

			// Add inline flag actions
			if hasInlineFlags {
				// Check for duplicates within flags
				if err := config.CheckDuplicatePaths(copyFlags, symlinkFlags); err != nil {
					return err
				}

				flagActions := config.BuildActionsFromFlags(copyFlags, symlinkFlags, runFlags)
				allActions = append(allActions, flagActions...)
			}

			// Execute all actions
			if len(allActions) > 0 {
				if err := setup.ExecuteActions(currentWorktreeRoot, targetPath, allActions); err != nil {
					return err
				}
				appliedActions = len(allActions)
			}
		}

		// 7. Check dirty status after config actions
		dirty, err := git.IsDirty(targetPath)
		if err != nil {
			return err
		}

		dirtyStr := "clean"
		if dirty {
			dirtyStr = "dirty"
		}

		// 8. Confirmation messages to stderr
		if appliedActions > 0 && configFileName != "" {
			fmt.Fprintf(os.Stderr, "Applied %s (%d actions)\n", configFileName, appliedActions)
		} else if appliedActions > 0 {
			fmt.Fprintf(os.Stderr, "Applied %d inline actions\n", appliedActions)
		}

		fmt.Fprintf(os.Stderr, "Switched to %s (branch: %s, %s)\n", basename, branchName, dirtyStr)

		// 9. Output path to stdout if flag is set
		if outputPath {
			fmt.Println(targetPath)
		} else {
			// Also print confirmation to stdout when not using --output-path
			fmt.Printf("Switched to %s (branch: %s, %s)\n", basename, branchName, dirtyStr)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVar(&configFlag, "config", "", "use .wtconfig-<name> instead of default .wtconfig")
	newCmd.Flags().BoolVar(&skipConfig, "skip-config", false, "skip all config actions")
	newCmd.Flags().StringSliceVar(&copyFlags, "copy", []string{}, "inline copy overrides (repeatable)")
	newCmd.Flags().StringSliceVar(&symlinkFlags, "symlink", []string{}, "inline symlink overrides (repeatable)")
	newCmd.Flags().StringSliceVar(&runFlags, "run", []string{}, "inline run commands (repeatable)")
}
