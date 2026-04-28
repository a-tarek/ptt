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
	copyEnvFlag    string
	setFlags       []string
	branchFlag     string
	prFlag         bool
	prTitleFlag    string
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

		// CurrentWorktreeRoot may fail from bare repo root; fall back to homePath
		currentWorktreeRoot, err := git.CurrentWorktreeRoot()
		if err != nil {
			currentWorktreeRoot = homePath
		}

		configRoot, err := git.ConfigRoot()
		if err != nil {
			return err
		}

		// 3. Validate name and compute target path. Worktree name becomes a directory
		// (and, when -b is not given, also the new branch name) — slashes break path
		// resolution and create awkward nested layouts. Hyphens are the only allowed
		// separator.
		name := args[0]
		if err := validateWorktreeName(name); err != nil {
			return err
		}
		targetPath, err := git.WorktreePath(homePath, name)
		if err != nil {
			return err
		}

		// 4. Determine branch name
		branchName := name
		if branchFlag != "" {
			branchName = branchFlag
		}

		// 5. Parse --set flags
		setOverrides, err := parseSetFlags(setFlags)
		if err != nil {
			return err
		}

		// 5b. Validate and build config actions (before creating worktree)
		var allActions []config.Action
		hasInlineFlags := len(copyFlags) > 0 || len(symlinkFlags) > 0 || len(runFlags) > 0
		hasCopyEnv := copyEnvFlag != "" || len(setFlags) > 0

		if !skipConfig || hasInlineFlags || hasCopyEnv {
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
					createActions, parseErr := parseCreateActions(configPath, currentWorktreeRoot)
					if parseErr != nil {
						return parseErr
					}
					allActions = append(allActions, createActions...)
				}
				// Silently skip if config file not found
			} else if !skipConfig && hasInlineFlags && configFlag != "" {
				// --config with inline flags: load named config, then append inline flags
				configPath, cfgErr := config.ResolveConfigPath(configRoot, configFlag)
				if cfgErr == nil {
					createActions, parseErr := parseCreateActions(configPath, currentWorktreeRoot)
					if parseErr != nil {
						return parseErr
					}
					allActions = append(allActions, createActions...)
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

			// Build inline --copy-env action
			if copyEnvFlag != "" {
				inlineCopyEnv := config.Action{
					Type: config.ActionCopyEnv,
					Path: copyEnvFlag,
					Line: 0,
					CopyEnv: &config.CopyEnvConfig{
						File:      copyEnvFlag,
						Vars:      map[string]config.EnvVarRule{},
						Overrides: setOverrides,
					},
				}
				allActions = append(allActions, inlineCopyEnv)
			}

			// Merge --set overrides into config-based copyEnv actions
			if len(setOverrides) > 0 {
				for i := range allActions {
					if allActions[i].Type == config.ActionCopyEnv && allActions[i].CopyEnv != nil {
						if allActions[i].CopyEnv.Overrides == nil {
							allActions[i].CopyEnv.Overrides = make(map[string]string)
						}
						for k, v := range setOverrides {
							allActions[i].CopyEnv.Overrides[k] = v
						}
					}
				}
			}

			// Validate copyEnv actions with merged overrides
			var copyEnvActions []config.Action
			for _, a := range allActions {
				if a.Type == config.ActionCopyEnv {
					copyEnvActions = append(copyEnvActions, a)
				}
			}
			if len(copyEnvActions) > 0 {
				if err := config.ValidateActions(currentWorktreeRoot, copyEnvActions); err != nil {
					return err
				}
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
		var createCmd *exec.Cmd
		if branchFlag != "" {
			// --branch: check out existing branch, don't create new
			createCmd = exec.Command("git", "worktree", "add", targetPath, branchName)
		} else {
			createCmd = exec.Command("git", "worktree", "add", targetPath, "-b", branchName)
		}
		output, err := createCmd.CombinedOutput()
		if err != nil {
			if branchFlag == "" {
				// Branch might already exist, try without -b
				createCmd = exec.Command("git", "worktree", "add", targetPath, branchName)
				output, err = createCmd.CombinedOutput()
			}
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

		// 10. Auto-PR (best-effort: failures print but don't roll back the worktree)
		if prFlag {
			if err := openDraftPR(targetPath, branchName, prTitleFlag); err != nil {
				fmt.Fprintf(os.Stderr, "\n--pr failed: %v\n", err)
				fmt.Fprintf(os.Stderr, "Worktree is intact; resume manually with: cd %s && git push -u origin %s && gh pr create --draft --fill\n", targetPath, branchName)
			}
		}

		// 11. Output path to stdout for shell wrapper
		if outputPath {
			fmt.Println(targetPath)
		} else {
			fmt.Fprintf(os.Stderr, "\nRun: ptt cd %s\n", name)
		}

		return nil
	},
}

// validateWorktreeName rejects names that would break ptt's flat sibling layout
// or create unintended branch namespaces.
func validateWorktreeName(name string) error {
	if name == "" {
		return fmt.Errorf("worktree name cannot be empty")
	}
	if strings.ContainsRune(name, '/') {
		return fmt.Errorf("worktree name %q contains '/': use hyphens (e.g. %q)", name, strings.ReplaceAll(name, "/", "-"))
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("worktree name %q cannot start with '-'", name)
	}
	return nil
}

// openDraftPR scaffolds a draft PR for the freshly-created worktree:
// 1. empty initial commit (so there's something to push)
// 2. push -u origin <branch>
// 3. gh pr create --draft --fill (or --title if prTitle is set)
func openDraftPR(worktreePath, branch, prTitle string) error {
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not found in PATH (install from https://cli.github.com/)")
	}

	commitMsg := branch
	if prTitle != "" {
		commitMsg = prTitle
	}
	commitCmd := exec.Command("git", "commit", "--allow-empty", "-m", commitMsg)
	commitCmd.Dir = worktreePath
	if out, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit --allow-empty: %s", strings.TrimSpace(string(out)))
	}

	pushCmd := exec.Command("git", "push", "-u", "origin", branch)
	pushCmd.Dir = worktreePath
	if out, err := pushCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push -u origin %s: %s", branch, strings.TrimSpace(string(out)))
	}

	prArgs := []string{"pr", "create", "--draft"}
	if prTitle != "" {
		prArgs = append(prArgs, "--title", prTitle, "--body", "")
	} else {
		prArgs = append(prArgs, "--fill")
	}
	prCmd := exec.Command("gh", prArgs...)
	prCmd.Dir = worktreePath
	out, err := prCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gh pr create: %s", strings.TrimSpace(string(out)))
	}
	fmt.Fprintf(os.Stderr, "%s", string(out))
	return nil
}

// parseCreateActions parses and validates create actions from a config file.
// Dispatches to YAML or text parser based on file extension.
func parseCreateActions(configPath, srcRoot string) ([]config.Action, error) {
	if config.IsYAMLConfig(configPath) {
		lc, err := config.ParseYAMLFile(configPath)
		if err != nil {
			return nil, err
		}
		if err := config.ValidateActions(srcRoot, lc.Create); err != nil {
			return nil, err
		}
		return lc.Create, nil
	}

	actions, err := config.ParseFile(configPath)
	if err != nil {
		return nil, err
	}
	if err := config.ValidateActions(srcRoot, actions); err != nil {
		return nil, err
	}
	return actions, nil
}

// parseSetFlags parses --set KEY=VALUE flags into a map.
func parseSetFlags(flags []string) (map[string]string, error) {
	result := make(map[string]string, len(flags))
	for _, flag := range flags {
		key, value, err := parseSetFlag(flag)
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

// parseSetFlag splits a KEY=VALUE string on the first '='.
func parseSetFlag(s string) (string, string, error) {
	eqIdx := strings.Index(s, "=")
	if eqIdx < 0 {
		return "", "", fmt.Errorf("invalid --set flag %q: expected KEY=VALUE", s)
	}
	key := s[:eqIdx]
	if key == "" {
		return "", "", fmt.Errorf("invalid --set flag %q: key cannot be empty", s)
	}
	return key, s[eqIdx+1:], nil
}

func init() {
	rootCmd.AddCommand(mkCmd)
	mkCmd.Flags().StringVar(&configFlag, "config", "", "use named config (.pttconfig/<name>)")
	mkCmd.Flags().BoolVar(&skipConfig, "skip-config", false, "skip all config actions")
	mkCmd.Flags().StringSliceVar(&copyFlags, "copy", []string{}, "inline copy overrides (repeatable)")
	mkCmd.Flags().StringSliceVar(&symlinkFlags, "symlink", []string{}, "inline symlink overrides (repeatable)")
	mkCmd.Flags().StringSliceVar(&runFlags, "run", []string{}, "inline run commands (repeatable)")
	mkCmd.Flags().StringVar(&copyEnvFlag, "copy-env", "", "copy env file with variable transformations")
	mkCmd.Flags().StringArrayVar(&setFlags, "set", []string{}, "set env var override (KEY=VALUE, repeatable)")
	mkCmd.Flags().StringVarP(&branchFlag, "branch", "b", "", "use existing branch instead of creating new")
	mkCmd.RegisterFlagCompletionFunc("branch", branchNameCompletion)
	mkCmd.Flags().BoolVar(&prFlag, "pr", false, "after creating the worktree, open a draft PR (empty commit + push + gh pr create)")
	mkCmd.Flags().StringVar(&prTitleFlag, "pr-title", "", "title for the draft PR (also used as the empty commit message); defaults to the branch name")
}
