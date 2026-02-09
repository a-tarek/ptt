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
	configFlag     string
	skipConfig     bool
	copyFlags      []string
	symlinkFlags   []string
	runFlags       []string
)

var mkCmd = &cobra.Command{
	Use:     "mk <name>",
	Aliases: []string{"new"},
	Short:   "Create a new worktree",
	Args:    cobra.ExactArgs(1),
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

		configRoot, err := git.ConfigRoot()
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

		// 5. Validate and build config actions (before creating worktree)
		var allActions []config.Action
		hasInlineFlags := len(copyFlags) > 0 || len(symlinkFlags) > 0 || len(runFlags) > 0

		if !skipConfig || hasInlineFlags {
			// Load file-based config (unless --skip-config or inline flags override)
			if !skipConfig && !hasInlineFlags {
				var configPath string
				if configFlag != "" {
					configPath, err = config.ResolveConfigPath(configRoot, configFlag)
				} else {
					configPath, err = config.ResolveConfigPath(configRoot, "")
				}

				// If config file found, parse and validate
				if err == nil {

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
				configPath, cfgErr := config.ResolveConfigPath(configRoot, configFlag)
				if cfgErr == nil {

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

			// Validate inline flags
			if hasInlineFlags {
				if err := config.CheckDuplicatePaths(copyFlags, symlinkFlags); err != nil {
					return err
				}

				flagActions := config.BuildActionsFromFlags(copyFlags, symlinkFlags, runFlags, os.Args)
				allActions = append(allActions, flagActions...)
			}
		}

		// 6. Build complete task list
		basename := filepath.Base(targetPath)
		tasks := ui.NewTaskList()
		tasks.Add("create:", basename)
		for _, action := range allActions {
			tasks.Add(action.Type+":", action.Path)
		}
		tasks.Add("cd:", basename)

		// 7. Create worktree (all validation passed)
		createCmd := exec.Command("git", "worktree", "add", targetPath, "-b", branchName)
		output, err := createCmd.CombinedOutput()
		if err != nil {
			// Branch might already exist, try without -b
			createCmd = exec.Command("git", "worktree", "add", targetPath, branchName)
			output, err = createCmd.CombinedOutput()
			if err != nil {
				tasks.FailRemaining(0)
				return fmt.Errorf("failed to create worktree: %s", strings.TrimSpace(string(output)))
			}
		}
		tasks.MarkDone(0)

		// 8. Execute config actions
		if len(allActions) > 0 {
			if err := setup.ExecuteActions(currentWorktreeRoot, targetPath, allActions, tasks, 1); err != nil {
				return err
			}
		}

		// 9. cd
		tasks.MarkDone(tasks.Len() - 1)

		// 10. Output path to stdout for shell wrapper
		if outputPath {
			fmt.Println(targetPath)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(mkCmd)
	mkCmd.Flags().StringVar(&configFlag, "config", "", "use named config (.pttconfig/<name>)")
	mkCmd.Flags().BoolVar(&skipConfig, "skip-config", false, "skip all config actions")
	mkCmd.Flags().StringSliceVar(&copyFlags, "copy", []string{}, "inline copy overrides (repeatable)")
	mkCmd.Flags().StringSliceVar(&symlinkFlags, "symlink", []string{}, "inline symlink overrides (repeatable)")
	mkCmd.Flags().StringSliceVar(&runFlags, "run", []string{}, "inline run commands (repeatable)")
}
