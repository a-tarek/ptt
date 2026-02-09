package initcmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/a-tarek/ptt/internal/git"
)

// RepoType represents the type of git repository
type RepoType int

const (
	RepoTypeNotGit RepoType = iota
	RepoTypeNormal
	RepoTypeRawBare
	RepoTypePttBare
)

// String returns human-readable description of the repo type
func (r RepoType) String() string {
	switch r {
	case RepoTypeNotGit:
		return "not a git repo"
	case RepoTypeNormal:
		return "standard clone"
	case RepoTypeRawBare:
		return "raw bare repo"
	case RepoTypePttBare:
		return "ptt-managed repo"
	default:
		return "unknown"
	}
}

// WorktreeInfo contains information about a worktree
type WorktreeInfo struct {
	Path   string
	Branch string
}

// RepoInfo contains detected repository information
type RepoInfo struct {
	RepoRoot            string
	DefaultBranch       string
	CurrentBranch       string
	HasRemote           bool
	RemoteURL           string
	ExistingWorktrees   []WorktreeInfo
	ContainerRoot       string
	IsBareFromWorktree  bool
}

// ProgressCallback is a function called with progress updates
type ProgressCallback func(step string)

// DetectRepoType determines the type of repository we're in
func DetectRepoType() (RepoType, *RepoInfo, error) {
	info := &RepoInfo{}

	// 1. Check if inside git repo
	if !git.IsInsideGitRepo() {
		return RepoTypeNotGit, info, nil
	}

	// 2. Check for ptt bare repo
	bareRoot, err := git.BareRepoRoot()
	if err == nil {
		info.ContainerRoot = bareRoot

		// Check if CWD is the container root itself
		cwd, err := os.Getwd()
		if err == nil {
			cwdAbs, _ := filepath.Abs(cwd)
			containerAbs, _ := filepath.Abs(bareRoot)
			if cwdAbs != containerAbs {
				info.IsBareFromWorktree = true
			}
		}

		// Populate info for ptt bare repo
		if err := populatePttBareInfo(info); err != nil {
			return RepoTypePttBare, info, err
		}

		return RepoTypePttBare, info, nil
	}

	// 3. Check for raw bare repo
	isBare, err := git.IsBareRepository()
	if err != nil {
		return RepoTypeNotGit, info, fmt.Errorf("failed to check if bare: %w", err)
	}

	if isBare {
		// Get bare repo root
		repoRoot, err := git.GetRepoRoot()
		if err != nil {
			return RepoTypeRawBare, info, fmt.Errorf("failed to get repo root: %w", err)
		}
		info.RepoRoot = repoRoot

		// Populate existing worktrees
		if err := populateWorktrees(info); err != nil {
			return RepoTypeRawBare, info, err
		}

		// Get default branch for bare repo
		if err := populateDefaultBranchForBare(info.RepoRoot, info); err != nil {
			return RepoTypeRawBare, info, err
		}

		// Get remote info
		populateRemoteInfo(info.RepoRoot, info)

		return RepoTypeRawBare, info, nil
	}

	// 4. Normal repository
	repoRoot, err := git.GetRepoRoot()
	if err != nil {
		return RepoTypeNormal, info, fmt.Errorf("failed to get repo root: %w", err)
	}
	info.RepoRoot = repoRoot

	// Get current branch
	currentBranch, err := git.CurrentBranch()
	if err == nil {
		info.CurrentBranch = currentBranch
	}

	// Get remote info first
	populateRemoteInfo(info.RepoRoot, info)

	// Determine default branch
	// 1. Try remote's HEAD (if has remote)
	if info.HasRemote {
		cmd := exec.Command("git", "-C", repoRoot, "symbolic-ref", "refs/remotes/origin/HEAD")
		output, err := cmd.Output()
		if err == nil {
			headRef := strings.TrimSpace(string(output))
			if strings.HasPrefix(headRef, "refs/remotes/origin/") {
				branch := strings.TrimPrefix(headRef, "refs/remotes/origin/")
				info.DefaultBranch = branch
			}
		}
	}

	// 2. Fallback to checking for main/master branches
	if info.DefaultBranch == "" {
		mainCmd := exec.Command("git", "-C", repoRoot, "show-ref", "--verify", "--quiet", "refs/heads/main")
		if mainCmd.Run() == nil {
			info.DefaultBranch = "main"
		} else {
			masterCmd := exec.Command("git", "-C", repoRoot, "show-ref", "--verify", "--quiet", "refs/heads/master")
			if masterCmd.Run() == nil {
				info.DefaultBranch = "master"
			}
		}
	}

	// 3. Final fallback to current branch
	if info.DefaultBranch == "" && info.CurrentBranch != "" {
		info.DefaultBranch = info.CurrentBranch
	}

	return RepoTypeNormal, info, nil
}

// populatePttBareInfo populates info for ptt bare repos
func populatePttBareInfo(info *RepoInfo) error {
	bareDir := filepath.Join(info.ContainerRoot, ".bare")

	// Get default branch
	if err := populateDefaultBranchForBare(bareDir, info); err != nil {
		return err
	}

	// Get remote info
	populateRemoteInfo(info.ContainerRoot, info)

	// Populate existing worktrees
	if err := populateWorktrees(info); err != nil {
		return err
	}

	return nil
}

// populateDefaultBranchForBare detects the default branch for a bare repo
func populateDefaultBranchForBare(bareDir string, info *RepoInfo) error {
	// Try symbolic-ref HEAD first
	cmd := exec.Command("git", "-C", bareDir, "symbolic-ref", "HEAD")
	output, err := cmd.Output()
	if err == nil {
		headRef := strings.TrimSpace(string(output))
		if strings.HasPrefix(headRef, "refs/heads/") {
			info.DefaultBranch = strings.TrimPrefix(headRef, "refs/heads/")
			return nil
		}
	}

	// Fallback: check for main then master
	mainCmd := exec.Command("git", "-C", bareDir, "show-ref", "--verify", "--quiet", "refs/heads/main")
	if mainCmd.Run() == nil {
		info.DefaultBranch = "main"
		return nil
	}

	masterCmd := exec.Command("git", "-C", bareDir, "show-ref", "--verify", "--quiet", "refs/heads/master")
	if masterCmd.Run() == nil {
		info.DefaultBranch = "master"
		return nil
	}

	return fmt.Errorf("no default branch found (checked main and master)")
}

// populateRemoteInfo gets remote information
func populateRemoteInfo(repoPath string, info *RepoInfo) {
	cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err == nil {
		info.HasRemote = true
		info.RemoteURL = strings.TrimSpace(string(output))
	}
}

// populateWorktrees gets existing worktrees
func populateWorktrees(info *RepoInfo) error {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var current WorktreeInfo
	info.ExistingWorktrees = []WorktreeInfo{}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			if current.Path != "" {
				info.ExistingWorktrees = append(info.ExistingWorktrees, current)
				current = WorktreeInfo{}
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			current.Path = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch ") {
			branchRef := strings.TrimPrefix(line, "branch ")
			if strings.HasPrefix(branchRef, "refs/heads/") {
				current.Branch = strings.TrimPrefix(branchRef, "refs/heads/")
			}
		} else if line == "bare" {
			// Skip bare entry (the bare repo itself)
			current = WorktreeInfo{}
		}
	}

	// Last entry if no trailing newline
	if current.Path != "" {
		info.ExistingWorktrees = append(info.ExistingWorktrees, current)
	}

	return nil
}
