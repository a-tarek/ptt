package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/a-tarek/ptt/internal/config"
	"github.com/a-tarek/ptt/internal/git"
	"github.com/a-tarek/ptt/internal/setup"
	"github.com/a-tarek/ptt/internal/ui"
	"github.com/spf13/cobra"
)

var mkFromPRCmd = &cobra.Command{
	Use:   "mk-from-pr <pr-number>",
	Short: "Create a worktree from a GitHub PR's head branch",
	Long: `Fetch a GitHub PR's head branch and create a worktree as a sibling of
the home worktree. Config actions are sourced from home, not from the
current worktree.

Steps:
  1. Resolve home worktree
  2. gh pr view <PR#> --json headRefName
  3. git fetch origin +refs/pull/<PR#>/head:refs/heads/<branch>
  4. git worktree add <home>/pr-<PR#> <branch>
  5. Run .pttconfig create actions sourced from home

Requires the 'gh' CLI in PATH and an authenticated GitHub session.`,
	Args: cobra.ExactArgs(1),
	RunE: mkFromPRImpl,
}

func init() {
	rootCmd.AddCommand(mkFromPRCmd)
}

func mkFromPRImpl(_ *cobra.Command, args []string) error {
	prNum, err := parsePRNumber(args[0])
	if err != nil {
		return err
	}

	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not found in PATH (install from https://cli.github.com/)")
	}

	if !git.IsInsideGitRepo() {
		return fmt.Errorf("not inside a git repository")
	}

	homePath, err := git.GetHomePath()
	if err != nil {
		return err
	}
	configRoot, err := git.ConfigRoot()
	if err != nil {
		return err
	}

	headRef, err := fetchPRHeadRef(homePath, prNum)
	if err != nil {
		return err
	}

	worktreeName := fmt.Sprintf("pr-%d", prNum)
	targetPath, err := git.WorktreePath(homePath, worktreeName)
	if err != nil {
		return err
	}

	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("worktree %q already exists at %s — remove it first with 'ptt rm %s'", worktreeName, targetPath, worktreeName)
	}

	refspec := fmt.Sprintf("+refs/pull/%d/head:refs/heads/%s", prNum, headRef)
	if out, err := runGitIn(homePath, "fetch", "origin", refspec); err != nil {
		return fmt.Errorf("git fetch %s failed: %s", refspec, strings.TrimSpace(out))
	}

	var actions []config.Action
	if cfgPath, cfgErr := config.ResolveConfigPath(configRoot, ""); cfgErr == nil {
		acts, parseErr := parseCreateActions(cfgPath, homePath)
		if parseErr != nil {
			return parseErr
		}
		actions = acts
	}

	basename := filepath.Base(targetPath)
	tasks := ui.NewTaskList()
	tasks.Add("create:", basename)
	for _, a := range actions {
		tasks.Add(a.Type+":", a.Path)
	}

	addCmd := exec.Command("git", "worktree", "add", targetPath, headRef)
	addCmd.Dir = homePath
	if out, err := addCmd.CombinedOutput(); err != nil {
		tasks.FailRemaining(0)
		return fmt.Errorf("git worktree add failed: %s", strings.TrimSpace(string(out)))
	}
	tasks.MarkDone(0)

	if len(actions) > 0 {
		if err := setup.ExecuteActions(homePath, targetPath, actions, tasks, 1); err != nil {
			return err
		}
	}

	fmt.Fprintf(os.Stderr, "\nRun: ptt cd %s\n", basename)
	return nil
}

func parsePRNumber(s string) (int, error) {
	n, err := strconv.Atoi(strings.TrimPrefix(s, "#"))
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid PR number %q (expected a positive integer)", s)
	}
	return n, nil
}

func fetchPRHeadRef(dir string, prNum int) (string, error) {
	cmd := exec.Command("gh", "pr", "view", strconv.Itoa(prNum), "--json", "headRefName")
	cmd.Dir = dir
	rawOut, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gh pr view %d failed: %s", prNum, strings.TrimSpace(string(rawOut)))
	}
	var meta struct {
		HeadRefName string `json:"headRefName"`
	}
	if err := json.Unmarshal(rawOut, &meta); err != nil {
		return "", fmt.Errorf("failed to parse gh output: %w", err)
	}
	if meta.HeadRefName == "" {
		return "", fmt.Errorf("PR #%d has no head ref", prNum)
	}
	return meta.HeadRefName, nil
}

// runGitIn runs `git <args...>` with cwd=dir and returns combined output.
func runGitIn(dir string, args ...string) (string, error) {
	c := exec.Command("git", args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	return string(out), err
}
