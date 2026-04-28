package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/a-tarek/ptt/internal/config"
	"github.com/a-tarek/ptt/internal/git"
	"github.com/a-tarek/ptt/internal/setup"
	"github.com/a-tarek/ptt/internal/staleness"
	"github.com/spf13/cobra"
)

var (
	forceDelete  bool
	deleteBranch bool
	rmRunFlags   []string
	rmConfigFlag string
	rmSkipConfig bool

	rmMergedBase string
	rmStaleArg   string
	rmGone       bool
	rmNoUpstream bool
	rmAssumeYes  bool
)

var rmCmd = &cobra.Command{
	Use:               "rm [worktree]",
	Aliases:           []string{"delete"},
	Short:             "Remove a worktree (or many, with --merged/--stale/--gone/--no-upstream)",
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: worktreeNameCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !git.IsInsideGitRepo() {
			return fmt.Errorf("not inside a git repository")
		}

		criteria, err := buildFilterCriteria(rmMergedBase, rmStaleArg, rmGone, rmNoUpstream)
		if err != nil {
			return err
		}

		// Mode A: filtered bulk delete (no name argument).
		if criteria.Empty() == false {
			if len(args) > 0 {
				return fmt.Errorf("pass either a worktree name OR filters, not both")
			}
			return runBulkDelete(criteria)
		}

		// Mode B: single delete (existing behavior).
		if len(args) == 0 {
			return fmt.Errorf("worktree name required (or pass a filter: --merged, --stale, --gone, --no-upstream)")
		}
		return runSingleDelete(args[0])
	},
}

// runSingleDelete is the original ptt rm <name> behavior.
func runSingleDelete(name string) error {
	wt, err := git.ResolveWorktree(name)
	if err != nil {
		return err
	}

	currentPath, _ := git.CurrentWorktreeRoot()
	if currentPath != "" && wt.Path == currentPath {
		return fmt.Errorf("can't delete current worktree")
	}

	dirty, dirtyErr := git.IsDirty(wt.Path)
	broken := dirtyErr != nil

	if dirty && !forceDelete && !broken {
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

	return removeOne(*wt, dirty, broken)
}

// runBulkDelete matches all worktrees against criteria, prompts once, then removes each.
// Always skips current, home, and (unless -f) dirty worktrees.
func runBulkDelete(criteria staleness.Criteria) error {
	worktrees, err := git.ListWorktrees()
	if err != nil {
		return err
	}

	currentPath, _ := git.CurrentWorktreeRoot()
	homePath, _ := git.GetHomePath()
	now := time.Now()

	type candidate struct {
		wt       git.Worktree
		decision staleness.Decision
		dirty    bool
		broken   bool
	}

	var matches []candidate
	var skippedDirty []string
	for _, wt := range worktrees {
		if wt.IsBare {
			continue
		}
		if wt.Path == currentPath {
			continue
		}
		if homePath != "" && wt.Path == homePath {
			continue
		}
		d := staleness.Evaluate(wt, criteria, now)
		if !d.Match {
			continue
		}
		dirty, dirtyErr := git.IsDirty(wt.Path)
		broken := dirtyErr != nil
		if dirty && !forceDelete {
			skippedDirty = append(skippedDirty, filepath.Base(wt.Path))
			continue
		}
		matches = append(matches, candidate{wt: wt, decision: d, dirty: dirty, broken: broken})
	}

	if len(matches) == 0 {
		fmt.Println("No worktrees match.")
		for _, s := range skippedDirty {
			fmt.Fprintf(os.Stderr, "  skipped (dirty): %s — pass -f to include\n", s)
		}
		if criteria.Gone {
			fmt.Println("Hint: run `git fetch --prune` to refresh remote tracking refs, then retry.")
		}
		return nil
	}

	// Preview
	fmt.Fprintln(os.Stderr, "The following worktrees will be removed:")
	for _, m := range matches {
		name := filepath.Base(m.wt.Path)
		lastEdited := "-"
		if !m.decision.LastEdited.IsZero() {
			lastEdited = humanAgo(now.Sub(m.decision.LastEdited))
		}
		dirtyTag := ""
		if m.dirty {
			dirtyTag = " (dirty)"
		}
		fmt.Fprintf(os.Stderr, "  %-30s %-25s %-12s %-10s %s%s\n",
			name, m.wt.Branch, m.decision.Upstream.String(), lastEdited,
			joinReasons(m.decision.Reasons), dirtyTag)
	}
	for _, s := range skippedDirty {
		fmt.Fprintf(os.Stderr, "  skipped (dirty): %s — pass -f to include\n", s)
	}
	if deleteBranch {
		fmt.Fprintln(os.Stderr, "Branches will also be deleted (-b).")
	}

	if !rmAssumeYes {
		fmt.Fprintf(os.Stderr, "\nProceed? [y/N] ")
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

	// Execute
	var failures []string
	for _, m := range matches {
		if err := removeOne(m.wt, m.dirty, m.broken); err != nil {
			failures = append(failures, fmt.Sprintf("  %s: %v", filepath.Base(m.wt.Path), err))
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("some worktrees failed to remove:\n%s", strings.Join(failures, "\n"))
	}
	return nil
}

// removeOne runs pre-remove hooks and removes a single worktree (and optionally its branch).
// Shared by single and bulk paths.
func removeOne(wt git.Worktree, dirty, broken bool) error {
	var removeActions []config.Action
	if !rmSkipConfig {
		configRoot, configErr := git.ConfigRoot()
		if configErr == nil {
			configPath, resolveErr := config.ResolveConfigPath(configRoot, rmConfigFlag)
			if resolveErr == nil && config.IsYAMLConfig(configPath) {
				lc, parseErr := config.ParseYAMLFile(configPath)
				if parseErr != nil {
					return parseErr
				}
				removeActions = append(removeActions, lc.Remove...)
			}
		}
	}
	for _, runCmd := range rmRunFlags {
		removeActions = append(removeActions, config.Action{Type: config.ActionRun, Path: runCmd})
	}
	if len(removeActions) > 0 {
		if err := config.ValidateRemoveActions(removeActions); err != nil {
			return err
		}
		if err := setup.RunPreRemoveHooks(wt.Path, removeActions, nil, 0); err != nil {
			if forceDelete {
				fmt.Fprintf(os.Stderr, "warning: pre-remove hook failed: %v\n", err)
			} else {
				return err
			}
		}
	}

	var removeCmd *exec.Cmd
	if dirty || broken {
		removeCmd = exec.Command("git", "worktree", "remove", "--force", wt.Path)
	} else {
		removeCmd = exec.Command("git", "worktree", "remove", wt.Path)
	}
	if output, err := removeCmd.CombinedOutput(); err != nil {
		if !broken {
			return fmt.Errorf("failed to remove worktree: %s", strings.TrimSpace(string(output)))
		}
		if err := os.RemoveAll(wt.Path); err != nil {
			return fmt.Errorf("failed to remove worktree directory: %w", err)
		}
		pruneCmd := exec.Command("git", "worktree", "prune")
		if pruneOut, err := pruneCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to prune worktrees: %s", strings.TrimSpace(string(pruneOut)))
		}
	}

	if deleteBranch && wt.Branch != "" {
		branchCmd := exec.Command("git", "branch", "-d", wt.Branch)
		if err := branchCmd.Run(); err != nil {
			if forceDelete {
				branchCmd = exec.Command("git", "branch", "-D", wt.Branch)
				if err := branchCmd.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to delete branch '%s'\n", wt.Branch)
				}
			} else {
				fmt.Fprintf(os.Stderr, "warning: branch '%s' not fully merged, use --force to delete\n", wt.Branch)
			}
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(rmCmd)
	rmCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "skip confirmation for dirty worktrees and force-delete unmerged branches")
	rmCmd.Flags().BoolVarP(&deleteBranch, "branch", "b", false, "also delete the branch after removing worktree")
	rmCmd.Flags().StringSliceVar(&rmRunFlags, "run", []string{}, "run command before deletion (repeatable)")
	rmCmd.Flags().StringVar(&rmConfigFlag, "config", "", "use named config (.pttconfig/<name>)")
	rmCmd.Flags().BoolVar(&rmSkipConfig, "skip-config", false, "skip config-based remove hooks")

	rmCmd.Flags().StringVar(&rmMergedBase, "merged", "", "bulk-remove worktrees merged into base (omit value to auto-detect main/master)")
	rmCmd.Flags().Lookup("merged").NoOptDefVal = "auto"
	rmCmd.Flags().StringVar(&rmStaleArg, "stale", "", "bulk-remove worktrees not edited in N (e.g. 30d, 2w, 3m; omit value for 30d)")
	rmCmd.Flags().Lookup("stale").NoOptDefVal = "30d"
	rmCmd.Flags().BoolVar(&rmGone, "gone", false, "bulk-remove worktrees whose upstream was deleted on remote")
	rmCmd.Flags().BoolVar(&rmNoUpstream, "no-upstream", false, "bulk-remove worktrees whose branch has no upstream")
	rmCmd.Flags().BoolVarP(&rmAssumeYes, "yes", "y", false, "skip confirmation prompt for bulk delete")
}
