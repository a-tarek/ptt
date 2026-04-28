package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/a-tarek/ptt/internal/git"
	"github.com/a-tarek/ptt/internal/staleness"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	showAllPaths bool
	lsMergedBase string
	lsStaleArg   string
	lsGone       bool
	lsNoUpstream bool
	lsShowPRs    bool
)

var lsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List all worktrees",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !git.IsInsideGitRepo() {
			return fmt.Errorf("not inside a git repository")
		}

		worktrees, err := git.ListWorktrees()
		if err != nil {
			return err
		}

		if len(worktrees) == 0 {
			return nil
		}

		var filtered []git.Worktree
		for _, wt := range worktrees {
			if !wt.IsBare {
				filtered = append(filtered, wt)
			}
		}
		worktrees = filtered

		if len(worktrees) == 0 {
			return nil
		}

		// Build filter criteria from flags. If anything is set, switch to filtered view.
		criteria, err := buildFilterCriteria(lsMergedBase, lsStaleArg, lsGone, lsNoUpstream)
		if err != nil {
			return err
		}

		if !criteria.Empty() || lsShowPRs {
			return renderFiltered(worktrees, criteria)
		}

		return renderDefault(worktrees)
	},
}

// buildFilterCriteria translates raw flag values into a staleness.Criteria.
// Resolves --merged "auto" sentinel to the detected default base.
// Parses --stale duration string (defaults to 30d when set without value via NoOptDefVal).
func buildFilterCriteria(mergedBase, staleArg string, gone, noUpstream bool) (staleness.Criteria, error) {
	c := staleness.Criteria{Gone: gone, NoUpstream: noUpstream}

	if mergedBase != "" {
		if mergedBase == "auto" {
			detected, err := git.DetectDefaultBase()
			if err != nil {
				return c, err
			}
			if detected == "" {
				return c, fmt.Errorf("could not auto-detect base branch (no main/master found); pass --merged <branch>")
			}
			c.MergedBase = detected
		} else {
			c.MergedBase = mergedBase
		}
	}

	if staleArg != "" {
		dur, err := staleness.ParseDuration(staleArg)
		if err != nil {
			return c, err
		}
		c.Stale = dur
	}

	return c, nil
}

// renderDefault emits the original `ptt ls` table format unchanged.
func renderDefault(worktrees []git.Worktree) error {
	currentPath, _ := git.CurrentWorktreeRoot()
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	normalColor := color.New(color.Reset)

	for _, wt := range worktrees {
		marker := " "
		if wt.Path == currentPath {
			marker = "*"
		}
		name := filepath.Base(wt.Path)
		dirty, _ := git.IsDirty(wt.Path)
		status := " "
		if dirty {
			status = "~"
		}

		if wt.Path == currentPath {
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
}

// renderFiltered shows worktrees with extra columns. When `c` is empty (no filter
// criteria), all worktrees are shown — useful when only --prs is set. Skips current
// and home worktrees in filtered mode for consistency with bulk-rm semantics, but
// when no filters are active, includes them so the user sees the full picture.
// Format: <name>  <branch>  <upstream>  <last-edited>  [<reasons>] [<pr-url>] [(dirty)]
func renderFiltered(worktrees []git.Worktree, c staleness.Criteria) error {
	now := time.Now()
	currentPath, _ := git.CurrentWorktreeRoot()
	yellow := color.New(color.FgYellow)
	dim := color.New(color.Faint)
	cyan := color.New(color.FgCyan)

	skipBoundary := !c.Empty()

	matched := 0
	for _, wt := range worktrees {
		if skipBoundary {
			if wt.Path == currentPath {
				continue
			}
			if isHomeWorktree(wt) {
				continue
			}
		}
		d := staleness.Evaluate(wt, c, now)
		if !d.Match {
			continue
		}
		matched++

		name := filepath.Base(wt.Path)
		dirty, _ := git.IsDirty(wt.Path)

		lastEdited := "-"
		if !d.LastEdited.IsZero() {
			lastEdited = humanAgo(now.Sub(d.LastEdited))
		}

		marker := " "
		if wt.Path == currentPath {
			marker = "*"
		}

		fmt.Printf("%s %-30s %-25s %-12s %-10s ", marker, name, wt.Branch, d.Upstream.String(), lastEdited)
		if reasons := joinReasons(d.Reasons); reasons != "" {
			dim.Printf("%s ", reasons)
		}
		if lsShowPRs {
			if url := lookupPRForBranch(wt.Branch); url != "" {
				cyan.Printf("%s ", url)
			}
		}
		if dirty {
			yellow.Printf("(dirty)")
		}
		fmt.Println()
	}

	if matched == 0 {
		// Only print "no matches" if filters were active (the no-filter --prs case
		// shouldn't say "no matches" — it just lists everything).
		if !c.Empty() {
			fmt.Println("No worktrees match.")
			if c.Gone {
				fmt.Println("Hint: run `git fetch --prune` to refresh remote tracking refs, then retry.")
			}
		}
	}

	return nil
}

// lookupPRForBranch returns the URL of an open PR whose head matches `branch`,
// or "" if none. Errors are silenced — a missing gh, no auth, or no PR all return "".
// Called once per worktree when --prs is on; sequential is fine for v1.
func lookupPRForBranch(branch string) string {
	if branch == "" {
		return ""
	}
	if _, err := exec.LookPath("gh"); err != nil {
		return ""
	}
	cmd := exec.Command("gh", "pr", "list", "--head", branch, "--json", "url", "--limit", "1", "--state", "all")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	var prs []struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(out, &prs); err != nil {
		return ""
	}
	if len(prs) == 0 {
		return ""
	}
	return prs[0].URL
}

func isHomeWorktree(wt git.Worktree) bool {
	homePath, err := git.GetHomePath()
	if err != nil {
		return false
	}
	return wt.Path == homePath
}

func humanAgo(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	day := 24 * time.Hour
	week := 7 * day
	month := 30 * day
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < day:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < week:
		return fmt.Sprintf("%dd ago", int(d/day))
	case d < month:
		return fmt.Sprintf("%dw ago", int(d/week))
	default:
		return fmt.Sprintf("%dmo ago", int(d/month))
	}
}

func joinReasons(rs []string) string {
	if len(rs) == 0 {
		return ""
	}
	out := "[" + rs[0]
	for _, r := range rs[1:] {
		out += "; " + r
	}
	return out + "]"
}

func init() {
	rootCmd.AddCommand(lsCmd)
	lsCmd.Flags().BoolVarP(&showAllPaths, "all", "a", false, "show full paths")

	lsCmd.Flags().StringVar(&lsMergedBase, "merged", "", "filter to branches merged into base (omit value to auto-detect main/master)")
	// Bare `--merged` (no value) becomes "auto", which we resolve later.
	lsCmd.Flags().Lookup("merged").NoOptDefVal = "auto"

	lsCmd.Flags().StringVar(&lsStaleArg, "stale", "", "filter to worktrees not edited in N (e.g. 30d, 2w, 3m; omit value for 30d)")
	lsCmd.Flags().Lookup("stale").NoOptDefVal = "30d"

	lsCmd.Flags().BoolVar(&lsGone, "gone", false, "filter to branches whose upstream was deleted on remote")
	lsCmd.Flags().BoolVar(&lsNoUpstream, "no-upstream", false, "filter to branches with no upstream configured")
	lsCmd.Flags().BoolVar(&lsShowPRs, "prs", false, "look up the PR URL for each worktree's branch (requires gh; one network call per worktree)")
}
