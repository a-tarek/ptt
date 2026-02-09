package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/a-tarek/ptt/internal/git"
	"github.com/a-tarek/ptt/internal/setup"
	"github.com/a-tarek/ptt/internal/ui"
	"github.com/spf13/cobra"
)

var mkBareRepoCmd = &cobra.Command{
	Use:   "mk-bare-repo",
	Short: "Convert normal clone to bare repo with nested worktrees",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Validate not in bare repo
		bareRoot, err := git.BareRepoRoot()
		if err == nil {
			return fmt.Errorf("already in a ptt bare repo: %s", bareRoot)
		}

		// 2. Validate inside git repo
		if !git.IsInsideGitRepo() {
			return fmt.Errorf("not inside a git repository")
		}

		// 3. Get remote origin URL
		remoteCmd := exec.Command("git", "remote", "get-url", "origin")
		remoteOutput, err := remoteCmd.Output()
		if err != nil {
			return fmt.Errorf("no remote origin configured -- mk-bare-repo requires a remote to clone from")
		}
		remoteURL := strings.TrimSpace(string(remoteOutput))

		// 4. Get current repo root
		currentRoot, err := git.CurrentWorktreeRoot()
		if err != nil {
			return fmt.Errorf("failed to get current repo root: %w", err)
		}

		// 5. Compute target directory
		parentDir := filepath.Dir(currentRoot)
		repoName := filepath.Base(currentRoot)
		targetDir := filepath.Join(parentDir, repoName+"-bare")

		if _, err := os.Stat(targetDir); err == nil {
			return fmt.Errorf("target directory already exists: %s", targetDir)
		}

		// 6. Build task list
		tasks := ui.NewTaskList()
		tasks.Add("clone:", ".bare")
		tasks.Add("create:", ".git pointer")
		tasks.Add("config:", "fetch refspec")
		tasks.Add("config:", "reflog")
		tasks.Add("fetch:", "origin")
		tasks.Add("worktree:", "default branch")

		// Check if source has .pttconfig
		hasPttConfig := false
		pttConfigSrc := filepath.Join(currentRoot, ".pttconfig")
		if _, err := os.Stat(pttConfigSrc); err == nil {
			hasPttConfig = true
			tasks.Add("copy:", ".pttconfig")
		}

		// Cleanup helper
		cleanup := func() {
			os.RemoveAll(targetDir)
		}

		// 7. Create container directory
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create target directory: %w", err)
		}

		// 8. Clone bare into .bare
		bareDir := filepath.Join(targetDir, ".bare")
		cloneCmd := exec.Command("git", "clone", "--bare", remoteURL, bareDir)
		if output, err := cloneCmd.CombinedOutput(); err != nil {
			cleanup()
			tasks.MarkFailed(0)
			tasks.FailRemaining(1)
			return fmt.Errorf("git clone --bare failed: %s", string(output))
		}
		tasks.MarkDone(0)

		// 9. Create .git pointer file
		gitPointer := filepath.Join(targetDir, ".git")
		if err := os.WriteFile(gitPointer, []byte("gitdir: ./.bare\n"), 0644); err != nil {
			cleanup()
			tasks.MarkFailed(1)
			tasks.FailRemaining(2)
			return fmt.Errorf("failed to create .git pointer: %w", err)
		}
		tasks.MarkDone(1)

		// 10. Configure fetch refspec
		configFetchCmd := exec.Command("git", "-C", targetDir, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
		if output, err := configFetchCmd.CombinedOutput(); err != nil {
			cleanup()
			tasks.MarkFailed(2)
			tasks.FailRemaining(3)
			return fmt.Errorf("git config fetch refspec failed: %s", string(output))
		}
		tasks.MarkDone(2)

		// 11. Enable reflog
		configReflogCmd := exec.Command("git", "-C", targetDir, "config", "core.logallrefupdates", "true")
		if output, err := configReflogCmd.CombinedOutput(); err != nil {
			cleanup()
			tasks.MarkFailed(3)
			tasks.FailRemaining(4)
			return fmt.Errorf("git config reflog failed: %s", string(output))
		}
		tasks.MarkDone(3)

		// 12. Fetch origin
		fetchCmd := exec.Command("git", "-C", targetDir, "fetch", "origin")
		if output, err := fetchCmd.CombinedOutput(); err != nil {
			cleanup()
			tasks.MarkFailed(4)
			tasks.FailRemaining(5)
			return fmt.Errorf("git fetch origin failed: %s", string(output))
		}
		tasks.MarkDone(4)

		// 13. Detect default branch
		var defaultBranch string
		mainCheckCmd := exec.Command("git", "-C", bareDir, "show-ref", "--verify", "--quiet", "refs/heads/main")
		if mainCheckCmd.Run() == nil {
			defaultBranch = "main"
		} else {
			masterCheckCmd := exec.Command("git", "-C", bareDir, "show-ref", "--verify", "--quiet", "refs/heads/master")
			if masterCheckCmd.Run() == nil {
				defaultBranch = "master"
			} else {
				cleanup()
				tasks.MarkFailed(5)
				tasks.FailRemaining(6)
				return fmt.Errorf("neither 'main' nor 'master' branch found")
			}
		}

		// 14. Create initial worktree
		worktreeName := strings.ReplaceAll(defaultBranch, "/", "-")
		worktreePath := filepath.Join(targetDir, worktreeName)
		worktreeCmd := exec.Command("git", "-C", bareDir, "worktree", "add", worktreePath, defaultBranch)
		if output, err := worktreeCmd.CombinedOutput(); err != nil {
			cleanup()
			tasks.MarkFailed(5)
			tasks.FailRemaining(6)
			return fmt.Errorf("git worktree add failed: %s", string(output))
		}
		tasks.MarkDone(5)

		// 15. Copy .pttconfig if exists
		if hasPttConfig {
			pttConfigDest := filepath.Join(targetDir, ".pttconfig")
			if err := setup.CopyPath(pttConfigSrc, pttConfigDest); err != nil {
				cleanup()
				tasks.MarkFailed(6)
				return fmt.Errorf("failed to copy .pttconfig: %w", err)
			}
			tasks.MarkDone(6)
		}

		// 16. Print success message
		fmt.Fprintf(os.Stderr, "\nBare repo ready: %s\n", targetDir)
		fmt.Fprintf(os.Stderr, "cd %s/%s && ptt mk <branch>\n", filepath.Base(targetDir), worktreeName)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(mkBareRepoCmd)
}
