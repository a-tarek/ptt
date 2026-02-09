package initcmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/a-tarek/ptt/internal/git"
)

// ValidationResult contains validation errors, warnings, and repair items
type ValidationResult struct {
	Errors      []string
	Warnings    []string
	RepairItems []string
}

// ValidateForInit validates the repository for initialization
func ValidateForInit(repoType RepoType, info *RepoInfo) (*ValidationResult, error) {
	result := &ValidationResult{
		Errors:      []string{},
		Warnings:    []string{},
		RepairItems: []string{},
	}

	switch repoType {
	case RepoTypeNotGit:
		result.Errors = append(result.Errors, "Not inside a git repository. Run 'git init' or 'git clone' first.")

	case RepoTypeNormal:
		// Check for dirty working tree
		isDirty, err := git.IsDirty(".")
		if err != nil {
			return result, err
		}
		if isDirty {
			result.Errors = append(result.Errors, "Working tree has uncommitted changes. Commit or stash your changes first.")
		}

		// Validate default branch detected
		if info.DefaultBranch == "" {
			result.Errors = append(result.Errors, "Could not detect default branch.")
		}

		// Warn if no remote
		if !info.HasRemote {
			result.Warnings = append(result.Warnings, "No remote origin configured. Bare repo will be local-only.")
		}

	case RepoTypeRawBare:
		// Validate default branch detected
		if info.DefaultBranch == "" {
			result.Errors = append(result.Errors, "Could not detect default branch.")
		}

		// Warn if no remote
		if !info.HasRemote {
			result.Warnings = append(result.Warnings, "No remote origin configured. Bare repo will be local-only.")
		}

	case RepoTypePttBare:
		// No blocking validation for ptt bare repos
		// Just detect repair items
		detectRepairItems(info, result)
	}

	return result, nil
}

// detectRepairItems checks for issues in ptt bare repos that need repair
func detectRepairItems(info *RepoInfo, result *ValidationResult) {
	if info.ContainerRoot == "" {
		return
	}

	// Check .bare/ directory exists
	bareDir := filepath.Join(info.ContainerRoot, ".bare")
	if _, err := os.Stat(bareDir); os.IsNotExist(err) {
		result.RepairItems = append(result.RepairItems, ".bare/ directory missing")
	}

	// Check .git file exists and contains correct content
	gitFile := filepath.Join(info.ContainerRoot, ".git")
	content, err := os.ReadFile(gitFile)
	if err != nil {
		result.RepairItems = append(result.RepairItems, ".git pointer file missing")
	} else if strings.TrimSpace(string(content)) != "gitdir: ./.bare" {
		result.RepairItems = append(result.RepairItems, ".git pointer file has incorrect content")
	}

	// Check fetch refspec (if remote exists)
	if info.HasRemote {
		cmd := exec.Command("git", "-C", info.ContainerRoot, "config", "remote.origin.fetch")
		output, err := cmd.Output()
		if err != nil {
			result.RepairItems = append(result.RepairItems, "fetch refspec not configured")
		} else {
			refspec := strings.TrimSpace(string(output))
			if refspec != "+refs/heads/*:refs/remotes/origin/*" {
				result.RepairItems = append(result.RepairItems, "fetch refspec incorrect")
			}
		}
	}

	// Check reflog enabled
	cmd := exec.Command("git", "-C", info.ContainerRoot, "config", "core.logallrefupdates")
	output, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(output)) != "true" {
		result.RepairItems = append(result.RepairItems, "reflog not enabled")
	}
}
