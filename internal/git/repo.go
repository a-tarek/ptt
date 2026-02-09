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

// GetHomePath returns the path of the first non-bare worktree (main checkout)
// For bare repos, returns the worktree matching the bare repo's HEAD branch
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
	var firstNonBareWorktree string
	var bareRepoPath string
	var bareRepoHEAD string

	// Parse worktree list to find bare repo and all worktrees
	type worktreeEntry struct {
		path   string
		branch string
		isBare bool
	}
	var worktrees []worktreeEntry
	var current worktreeEntry

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "worktree ") {
			// Save previous entry if exists
			if current.path != "" {
				worktrees = append(worktrees, current)
			}
			// Start new entry
			current = worktreeEntry{path: strings.TrimPrefix(line, "worktree ")}
		} else if strings.HasPrefix(line, "HEAD ") {
			current.branch = "" // HEAD without branch means detached
		} else if strings.HasPrefix(line, "branch ") {
			branchRef := strings.TrimPrefix(line, "branch ")
			// Extract branch name from refs/heads/branchname
			if strings.HasPrefix(branchRef, "refs/heads/") {
				current.branch = strings.TrimPrefix(branchRef, "refs/heads/")
			}
		} else if line == "bare" {
			current.isBare = true
		} else if line == "" || i == len(lines)-1 {
			// End of entry or end of output
			if current.path != "" {
				worktrees = append(worktrees, current)
				current = worktreeEntry{}
			}
		}
	}

	// Find bare repo and its HEAD
	for _, wt := range worktrees {
		if wt.isBare {
			bareRepoPath = wt.path
			break
		}
	}

	// If we found a bare repo, determine which branch its HEAD points to
	if bareRepoPath != "" {
		headCmd := exec.Command("git", "-C", bareRepoPath, "symbolic-ref", "HEAD")
		headOutput, err := headCmd.Output()
		if err == nil {
			headRef := strings.TrimSpace(string(headOutput))
			if strings.HasPrefix(headRef, "refs/heads/") {
				bareRepoHEAD = strings.TrimPrefix(headRef, "refs/heads/")
			}
		}

		// Find the worktree that matches the bare repo's HEAD branch
		if bareRepoHEAD != "" {
			for _, wt := range worktrees {
				if !wt.isBare && wt.branch == bareRepoHEAD {
					return wt.path, nil
				}
			}
		}
	}

	// Fallback: return first non-bare worktree
	for _, wt := range worktrees {
		if !wt.isBare {
			if firstNonBareWorktree == "" {
				firstNonBareWorktree = wt.path
			}
			// If no bare repo was found, the first worktree is the home
			if bareRepoPath == "" {
				return wt.path, nil
			}
		}
	}

	if firstNonBareWorktree != "" {
		return firstNonBareWorktree, nil
	}

	return "", fmt.Errorf("no non-bare worktree found")
}

// WorktreePath computes the target path for a new worktree
// For ptt bare repos: nested mode (e.g., /code/project-bare/staging)
// For standard bare repos: nested mode (e.g., /code/repo.git/staging)
// For regular repos: sibling mode (e.g., /code/wt-staging)
// Returns error if the computed path already exists
func WorktreePath(repoRoot string, name string) (string, error) {
	var targetPath string

	// First, check if we're in a ptt bare repo context
	bareRoot, err := BareRepoRoot()
	if err == nil {
		// Ptt bare repo mode: nested path under container root
		targetPath = filepath.Join(bareRoot, name)
	} else {
		// Fall back to legacy bare detection for backward compatibility
		// Check if repo root ends with .git (standard bare repo convention)
		if strings.HasSuffix(repoRoot, ".git") {
			// Standard bare repo: nested mode under parent directory
			parentDir := filepath.Dir(repoRoot)
			targetPath = filepath.Join(parentDir, name)
		} else {
			// Check worktree list for bare marker (another way to detect bare repos)
			cmd := exec.Command("git", "worktree", "list", "--porcelain")
			output, err := cmd.Output()
			isBare := false
			if err == nil {
				lines := strings.Split(string(output), "\n")
				// Check first worktree entry for bare marker
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "bare" {
						isBare = true
						break
					}
					// Stop at first empty line (end of first entry)
					if line == "" {
						break
					}
				}
			}

			if isBare {
				// Bare repo detected via worktree list: nested mode
				parentDir := filepath.Dir(repoRoot)
				targetPath = filepath.Join(parentDir, name)
			} else {
				// Non-bare mode: sibling path using repoRoot
				parentDir := filepath.Dir(repoRoot)
				repoName := filepath.Base(repoRoot)
				targetPath = filepath.Join(parentDir, repoName+"-"+name)
			}
		}
	}

	// Check if path already exists
	if _, err := os.Stat(targetPath); err == nil {
		return "", fmt.Errorf("path already exists: %s", targetPath)
	}

	return targetPath, nil
}

// BareRepoRoot returns the container root path when CWD is inside a ptt bare repo
// Returns error if not in a ptt bare repo structure
//
// A ptt bare repo structure looks like:
//   project-bare/
//     .git          (file containing "gitdir: ./.bare")
//     .bare/        (actual bare repository)
//     main/         (worktree)
//     feature/      (worktree)
//
// This function returns the "project-bare" path when called from any worktree or the root
func BareRepoRoot() (string, error) {
	// Get the git common directory (shared git data directory)
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git rev-parse failed: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to execute git: %w", err)
	}

	commonDir := strings.TrimSpace(string(output))

	// Make path absolute if it's relative
	if !filepath.IsAbs(commonDir) {
		absCommonDir, err := filepath.Abs(commonDir)
		if err != nil {
			return "", fmt.Errorf("failed to resolve absolute path: %w", err)
		}
		commonDir = absCommonDir
	}

	// Check if the basename is ".bare" (ptt convention)
	if filepath.Base(commonDir) != ".bare" {
		return "", fmt.Errorf("not in ptt bare repo: common dir is not .bare")
	}

	// Get the container root (parent of .bare)
	containerRoot := filepath.Dir(commonDir)

	// Verify that containerRoot/.git exists and is a file (not directory)
	gitPath := filepath.Join(containerRoot, ".git")
	info, err := os.Lstat(gitPath)
	if err != nil {
		return "", fmt.Errorf("not in ptt bare repo: .git file not found at container root")
	}

	if info.IsDir() {
		return "", fmt.Errorf("not in ptt bare repo: .git is a directory, not a file")
	}

	return containerRoot, nil
}

// ConfigRoot returns the root directory where ptt config should be stored
// In bare repo context: returns the bare repo container root
// In non-bare context: returns the home worktree path
func ConfigRoot() (string, error) {
	// Try bare repo first
	bareRoot, err := BareRepoRoot()
	if err == nil {
		return bareRoot, nil
	}

	// Fall back to home path for non-bare repos
	return GetHomePath()
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
