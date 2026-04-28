package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/a-tarek/ptt/internal/git"
	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone <git-url> [name]",
	Short: "Clone a remote repo into ptt's bare-worktree layout",
	Long: `Clone a remote repository as a bare repo and set up the ptt layout:

  <name>/
  ├── .bare/          # the bare clone
  ├── .git            # pointer file (gitdir: ./.bare)
  ├── .pttconfig/     # placeholder for config
  └── main/           # initial worktree on the default branch

If [name] is omitted, the directory name is derived from the repo URL
(e.g. github.com/foo/bar → "bar").

After cloning, cd into <name>/main to start working, or run other ptt
commands from anywhere inside the layout.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: cloneImpl,
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}

func cloneImpl(_ *cobra.Command, args []string) error {
	gitURL := args[0]

	var name string
	if len(args) >= 2 {
		name = args[1]
	} else {
		var err error
		name, err = deriveCloneName(gitURL)
		if err != nil {
			return err
		}
	}

	if err := validateWorktreeName(name); err != nil {
		return fmt.Errorf("invalid clone name: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getcwd: %w", err)
	}
	containerRoot := filepath.Join(cwd, name)

	if _, err := os.Stat(containerRoot); err == nil {
		return fmt.Errorf("path %s already exists", containerRoot)
	}

	if err := os.MkdirAll(containerRoot, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", containerRoot, err)
	}

	rollback := func() {
		os.RemoveAll(containerRoot)
	}

	bareDir := filepath.Join(containerRoot, ".bare")
	cloneCmd := exec.Command("git", "clone", "--bare", gitURL, bareDir)
	cloneCmd.Stdout = os.Stderr
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		rollback()
		return fmt.Errorf("git clone --bare failed: %w", err)
	}

	gitFile := filepath.Join(containerRoot, ".git")
	if err := os.WriteFile(gitFile, []byte("gitdir: ./.bare\n"), 0644); err != nil {
		rollback()
		return fmt.Errorf("write .git pointer: %w", err)
	}

	// Configure remote tracking refspec so future fetches populate refs/remotes/origin/*.
	if out, err := runGitIn(containerRoot, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*"); err != nil {
		rollback()
		return fmt.Errorf("git config remote.origin.fetch: %s", strings.TrimSpace(out))
	}
	if out, err := runGitIn(containerRoot, "fetch", "origin"); err != nil {
		// Non-fatal: we still have the bare clone and can recover. Print and continue.
		fmt.Fprintf(os.Stderr, "warning: initial fetch failed: %s\n", strings.TrimSpace(out))
	}

	// Detect default branch from the bare clone (HEAD's symbolic ref).
	defaultBranch, err := detectCloneDefaultBranch(bareDir)
	if err != nil {
		rollback()
		return err
	}

	// Create the initial main worktree as a sibling of .bare.
	mainPath := filepath.Join(containerRoot, "main")
	if out, err := runGitIn(containerRoot, "worktree", "add", mainPath, defaultBranch); err != nil {
		rollback()
		return fmt.Errorf("create initial worktree: %s", strings.TrimSpace(out))
	}

	// Drop a placeholder .pttconfig directory so users can drop a default.yml in later.
	pttConfigDir := filepath.Join(containerRoot, ".pttconfig")
	_ = os.MkdirAll(pttConfigDir, 0755)

	if outputPath {
		fmt.Println(mainPath)
	} else {
		fmt.Fprintf(os.Stderr, "\nCloned %s into %s\nRun: cd %s\n", gitURL, containerRoot, mainPath)
	}
	return nil
}

// deriveCloneName extracts a sensible directory name from a clone URL.
// Handles github SSH (git@github.com:foo/bar.git) and HTTPS (https://github.com/foo/bar.git) forms.
func deriveCloneName(gitURL string) (string, error) {
	trimmed := strings.TrimSuffix(gitURL, ".git")

	// SCP-like SSH: git@host:path
	if idx := strings.Index(trimmed, ":"); idx > 0 && !strings.HasPrefix(trimmed, "http") && !strings.HasPrefix(trimmed, "ssh://") {
		// Only treat as SCP if there's no scheme separator; "://" is a URL.
		if !strings.Contains(trimmed[:idx], "/") {
			path := trimmed[idx+1:]
			return filepath.Base(path), nil
		}
	}

	if strings.Contains(trimmed, "://") {
		u, err := url.Parse(trimmed)
		if err != nil {
			return "", fmt.Errorf("could not parse URL %q: %w", gitURL, err)
		}
		return filepath.Base(u.Path), nil
	}

	// Fallback: just take the last path segment.
	return filepath.Base(trimmed), nil
}

// detectCloneDefaultBranch returns the default branch name of a bare clone by
// reading HEAD's symbolic ref (which `git clone --bare` sets to track the remote HEAD).
func detectCloneDefaultBranch(bareDir string) (string, error) {
	cmd := exec.Command("git", "-C", bareDir, "symbolic-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		// Fall back to common defaults if HEAD is missing/unreadable.
		if base, _ := git.DetectDefaultBase(); base != "" {
			return base, nil
		}
		return "", fmt.Errorf("could not detect default branch from %s: %w", bareDir, err)
	}
	ref := strings.TrimSpace(string(out))
	return strings.TrimPrefix(ref, "refs/heads/"), nil
}
