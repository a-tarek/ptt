package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// UpstreamStatus describes a branch's relationship to its remote tracking ref.
type UpstreamStatus int

const (
	// UpstreamUnknown means we couldn't determine the status (branch missing, detached, etc.).
	UpstreamUnknown UpstreamStatus = iota
	// UpstreamActive means the branch has an upstream and that upstream still exists.
	UpstreamActive
	// UpstreamGone means the branch had an upstream that has been deleted on the remote.
	UpstreamGone
	// UpstreamNone means the branch was never configured with an upstream.
	UpstreamNone
)

// String returns the lowercase canonical name of the status.
func (s UpstreamStatus) String() string {
	switch s {
	case UpstreamActive:
		return "active"
	case UpstreamGone:
		return "gone"
	case UpstreamNone:
		return "no-upstream"
	default:
		return "unknown"
	}
}

// BranchUpstreamStatus reports the upstream status of a branch by name.
// Detached HEAD or missing branch returns UpstreamUnknown.
func BranchUpstreamStatus(branch string) (UpstreamStatus, error) {
	if branch == "" {
		return UpstreamUnknown, nil
	}

	// Check if upstream is configured at all.
	upstreamCmd := exec.Command("git", "rev-parse", "--abbrev-ref", branch+"@{upstream}")
	if err := upstreamCmd.Run(); err != nil {
		// Exit code 128 = no upstream configured. Treat as NoUpstream regardless of error text.
		return UpstreamNone, nil
	}

	// Upstream is configured. Check if it still exists with `for-each-ref`'s upstream:track token.
	// %(upstream:track) prints "[gone]" when the upstream ref is missing locally.
	trackCmd := exec.Command("git", "for-each-ref", "--format=%(upstream:track)", "refs/heads/"+branch)
	out, err := trackCmd.Output()
	if err != nil {
		return UpstreamUnknown, fmt.Errorf("git for-each-ref failed: %w", err)
	}
	if strings.Contains(string(out), "gone") {
		return UpstreamGone, nil
	}
	return UpstreamActive, nil
}

// IsMergedInto returns true if `branch` is fully merged into `base`.
// Uses `git merge-base --is-ancestor`: branch is merged iff its tip is an ancestor of base's tip.
// Note: this does NOT detect squash-merges (where a new commit replaces the branch's history).
func IsMergedInto(branch, base string) (bool, error) {
	if branch == "" || base == "" {
		return false, nil
	}
	if branch == base {
		return false, nil
	}
	cmd := exec.Command("git", "merge-base", "--is-ancestor", branch, base)
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		// Exit code 1 = not an ancestor. Anything else is a real error.
		if exitErr.ExitCode() == 1 {
			return false, nil
		}
	}
	return false, fmt.Errorf("git merge-base failed: %w", err)
}

// HeadModTime returns the mtime of the worktree's HEAD file.
// Reflects the last commit/checkout in that worktree — cheap and intent-aligned.
// For linked worktrees, HEAD lives at .git (a file pointing to gitdir/HEAD), so we follow it.
func HeadModTime(worktreePath string) (time.Time, error) {
	gitPath := filepath.Join(worktreePath, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return time.Time{}, fmt.Errorf("stat %s: %w", gitPath, err)
	}

	// Linked worktree: .git is a file like "gitdir: /abs/path/to/.bare/worktrees/<name>"
	if !info.IsDir() {
		data, err := os.ReadFile(gitPath)
		if err != nil {
			return time.Time{}, fmt.Errorf("read %s: %w", gitPath, err)
		}
		gitdir := strings.TrimSpace(strings.TrimPrefix(string(data), "gitdir:"))
		if !filepath.IsAbs(gitdir) {
			gitdir = filepath.Join(worktreePath, gitdir)
		}
		headPath := filepath.Join(gitdir, "HEAD")
		hi, err := os.Stat(headPath)
		if err != nil {
			return time.Time{}, fmt.Errorf("stat %s: %w", headPath, err)
		}
		return hi.ModTime(), nil
	}

	// Regular repo: HEAD is at .git/HEAD
	headPath := filepath.Join(gitPath, "HEAD")
	hi, err := os.Stat(headPath)
	if err != nil {
		return time.Time{}, fmt.Errorf("stat %s: %w", headPath, err)
	}
	return hi.ModTime(), nil
}

// DetectDefaultBase returns the first locally-existing branch from a candidate list,
// preferring 'main' then 'master'. Returns empty string and no error if neither exists —
// callers should treat that as "no default base detected" and surface a useful message.
func DetectDefaultBase() (string, error) {
	candidates := []string{"main", "master"}
	for _, c := range candidates {
		cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+c)
		if err := cmd.Run(); err == nil {
			return c, nil
		}
	}
	return "", nil
}
