package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsBareRepository returns true if the current repository is bare
func IsBareRepository() (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--is-bare-repository")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return false, fmt.Errorf("git rev-parse failed: %s", string(exitErr.Stderr))
		}
		return false, fmt.Errorf("failed to execute git: %w", err)
	}

	result := strings.TrimSpace(string(output))
	return result == "true", nil
}

// GetRepoRoot returns the repository root path
// For bare repos: returns the bare repository root (.git directory)
// For regular repos: returns the main checkout root
func GetRepoRoot() (string, error) {
	isBare, err := IsBareRepository()
	if err != nil {
		return "", err
	}

	if isBare {
		// For bare repos, --git-dir returns the bare repo root
		cmd := exec.Command("git", "rev-parse", "--git-dir")
		output, err := cmd.Output()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return "", fmt.Errorf("git rev-parse failed: %s", string(exitErr.Stderr))
			}
			return "", fmt.Errorf("failed to execute git: %w", err)
		}
		return strings.TrimSpace(string(output)), nil
	}

	// For regular repos, --show-toplevel returns the main checkout root
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

// GetHomePath returns the path of the first worktree (main checkout or bare repo root)
func GetHomePath() (string, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git worktree list failed: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to execute git: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "worktree ") {
			return strings.TrimPrefix(line, "worktree "), nil
		}
	}

	return "", fmt.Errorf("no worktree found")
}

// WorktreePath computes the target path for a new worktree
// For bare repos: nested mode (e.g., /code/wt/staging)
// For regular repos: sibling mode (e.g., /code/wt-staging)
// Returns error if the computed path already exists
func WorktreePath(repoRoot string, name string) (string, error) {
	isBare, err := IsBareRepository()
	if err != nil {
		return "", err
	}

	var targetPath string
	if isBare {
		// Nested mode: worktree under bare repo root
		targetPath = filepath.Join(repoRoot, name)
	} else {
		// Sibling mode: worktree as sibling to main checkout
		parentDir := filepath.Dir(repoRoot)
		repoName := filepath.Base(repoRoot)
		targetPath = filepath.Join(parentDir, repoName+"-"+name)
	}

	// Check if path already exists
	if _, err := os.Stat(targetPath); err == nil {
		return "", fmt.Errorf("path already exists: %s", targetPath)
	}

	return targetPath, nil
}

// CurrentBranch returns the current branch name
// Returns error if in detached HEAD state
func CurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git branch failed: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to execute git: %w", err)
	}

	branch := strings.TrimSpace(string(output))
	if branch == "" {
		return "", fmt.Errorf("detached HEAD state")
	}

	return branch, nil
}
