package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Worktree represents a git worktree
type Worktree struct {
	Path       string
	Branch     string
	IsBare     bool
	IsDetached bool
}

// ListWorktrees returns all worktrees in the current repository
func ListWorktrees() ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git worktree list failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute git: %w", err)
	}

	return parseWorktreePorcelain(string(output)), nil
}

func parseWorktreePorcelain(output string) []Worktree {
	var worktrees []Worktree
	var current Worktree

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			// Empty line separates worktrees
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = Worktree{}
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			current.Path = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch ") {
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		} else if strings.HasPrefix(line, "bare") {
			current.IsBare = true
		} else if strings.HasPrefix(line, "detached") {
			current.IsDetached = true
		}
	}

	// Last entry (no trailing newline)
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees
}

// IsDirty returns true if the worktree has uncommitted changes
func IsDirty(path string) (bool, error) {
	cmd := exec.Command("git", "-C", path, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return false, fmt.Errorf("git status failed: %s", string(exitErr.Stderr))
		}
		return false, fmt.Errorf("failed to execute git: %w", err)
	}

	// Check for actual changes (not just untracked files)
	// Porcelain format: XY filename
	// Untracked files start with "??"
	// Modified/staged files have other status codes (M, A, D, R, C, etc.)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// If line doesn't start with "??" it's a real change
		if !strings.HasPrefix(line, "?? ") {
			return true, nil
		}
	}

	return false, nil
}

// CurrentWorktreeRoot returns the root path of the current worktree
func CurrentWorktreeRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git rev-parse failed: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to execute git: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// IsInsideGitRepo returns true if currently inside a git repository
func IsInsideGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}
