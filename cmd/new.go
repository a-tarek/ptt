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
	"github.com/fatih/color"
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

		// 5. Validate and build config actions (before creating worktree)
		var allActions []config.Action
		hasInlineFlags := len(copyFlags) > 0 || len(symlinkFlags) > 0 || len(runFlags) > 0

		if !skipConfig || hasInlineFlags {
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

		// 6. Create worktree (all validation passed)
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

		// 7. Print status and execute actions
		basename := filepath.Base(targetPath)
		label := color.New(color.FgCyan)
		fmt.Fprintf(os.Stderr, "- %s %s\n", label.Sprint("create:"), basename)

		if len(allActions) > 0 {
			if err := setup.ExecuteActions(currentWorktreeRoot, targetPath, allActions); err != nil {
				return err
			}
		}

		fmt.Fprintf(os.Stderr, "- %s %s\n", label.Sprint("cd:"), basename)

		// 8. Output path to stdout for shell wrapper
		if outputPath {
			fmt.Println(targetPath)
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
