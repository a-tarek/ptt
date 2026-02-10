package initcmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// defaultYAMLConfig is the commented-out YAML template written to .pttconfig/default.yml on init.
var defaultYAMLConfig = []byte(`# ptt worktree config
# Uncomment and edit the actions below.
#
# create:
#   - copy: .env
#   - symlink: node_modules
#   - run: npm install
#   - copyEnv:
#       file: .env
#       vars:
#         PORT:
#           strategy: ptt_increment
#         BRANCH:
#           strategy: git branch --show-current
#
# remove:
#   - run: echo "cleaning up"
`)

// createDefaultYAMLConfig creates .pttconfig/default.yml with a commented-out YAML template.
func createDefaultYAMLConfig(pttconfigDir string) error {
	if err := os.MkdirAll(pttconfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create .pttconfig directory: %w", err)
	}
	defaultConfig := filepath.Join(pttconfigDir, "default.yml")
	if err := os.WriteFile(defaultConfig, defaultYAMLConfig, 0644); err != nil {
		return fmt.Errorf("failed to create .pttconfig/default.yml: %w", err)
	}
	return nil
}

// RestructureNormalRepo converts a normal git clone into a ptt-managed bare repo in place.
// This preserves untracked files by moving them through staging and into the final worktree.
//
// Algorithm:
// 1. Identify untracked files
// 2. Move all content (except .git, .pttconfig) to _ptt_staging/
// 3. Rename .git to .bare
// 4. Create .git pointer file
// 5. Set core.bare = true
// 6. Configure fetch refspec (if has remote)
// 7. Enable reflog
// 8. Fetch from remote (if has remote, best-effort)
// 9. Create default branch worktree
// 10. Copy untracked files from _ptt_staging into worktree
// 11. Create feature branch worktree (if current branch != default branch)
// 12. Clean up _ptt_staging
// 13. Create .pttconfig/default
func RestructureNormalRepo(repoRoot string, info *RepoInfo, progress ProgressCallback) error {
	// Build cleanup stack for rollback on failure
	var cleanupStack []func() error
	defer func() {
		if len(cleanupStack) > 0 {
			// Execute cleanup in reverse order
			for i := len(cleanupStack) - 1; i >= 0; i-- {
				_ = cleanupStack[i]() // Best effort cleanup
			}
		}
	}()

	stagingDir := filepath.Join(repoRoot, "_ptt_staging")

	// Step 1: Check if staging directory already exists
	if _, err := os.Stat(stagingDir); err == nil {
		return fmt.Errorf("staging directory already exists: %s (previous init may have failed, please clean up manually)", stagingDir)
	}

	// Step 2: Identify untracked files
	progress("Identifying untracked files")
	cmd := exec.Command("git", "-C", repoRoot, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get git status: %w", err)
	}

	var untrackedFiles []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "?? ") {
			// Remove "?? " prefix
			untrackedFiles = append(untrackedFiles, strings.TrimPrefix(line, "?? "))
		}
	}

	// Step 3: Move all content to staging directory
	progress("Moving content to staging")
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}
	cleanupStack = append(cleanupStack, func() error {
		return os.RemoveAll(stagingDir)
	})

	entries, err := os.ReadDir(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to read repo root: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		// Skip .git, .pttconfig, and the staging directory itself
		if name == ".git" || name == ".pttconfig" || name == "_ptt_staging" {
			continue
		}

		src := filepath.Join(repoRoot, name)
		dst := filepath.Join(stagingDir, name)

		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("failed to move %s to staging: %w", name, err)
		}
	}

	// Step 4: Rename .git to .bare
	progress("Restructuring git directory")
	gitDir := filepath.Join(repoRoot, ".git")
	bareDir := filepath.Join(repoRoot, ".bare")

	if err := os.Rename(gitDir, bareDir); err != nil {
		return fmt.Errorf("failed to rename .git to .bare: %w", err)
	}
	cleanupStack = append(cleanupStack, func() error {
		// Move .bare back to .git
		_ = os.Rename(bareDir, gitDir)
		// Move staging contents back
		entries, _ := os.ReadDir(stagingDir)
		for _, entry := range entries {
			src := filepath.Join(stagingDir, entry.Name())
			dst := filepath.Join(repoRoot, entry.Name())
			_ = os.Rename(src, dst)
		}
		return nil
	})

	// Step 5: Create .git pointer file
	progress("Creating .git pointer file")
	gitPointer := filepath.Join(repoRoot, ".git")
	if err := os.WriteFile(gitPointer, []byte("gitdir: ./.bare\n"), 0644); err != nil {
		return fmt.Errorf("failed to create .git pointer file: %w", err)
	}

	// Step 6: Set core.bare = true
	progress("Configuring bare repository")
	cmd = exec.Command("git", "-C", repoRoot, "config", "core.bare", "true")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set core.bare: %w", err)
	}

	// Step 7: Configure fetch refspec (if has remote)
	if info.HasRemote {
		progress("Configuring fetch refspec")
		cmd = exec.Command("git", "-C", repoRoot, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to configure fetch refspec: %w", err)
		}
	}

	// Step 8: Enable reflog
	progress("Enabling reflog")
	cmd = exec.Command("git", "-C", repoRoot, "config", "core.logallrefupdates", "true")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable reflog: %w", err)
	}

	// Step 9: Fetch from remote (if has remote, best-effort)
	if info.HasRemote {
		progress("Fetching from remote")
		cmd = exec.Command("git", "-C", repoRoot, "fetch", "origin")
		// Ignore errors - network issues shouldn't block restructuring
		_ = cmd.Run()
	}

	// Step 10: Create default branch worktree
	progress(fmt.Sprintf("Creating worktree for %s", info.DefaultBranch))
	worktreePath := filepath.Join(repoRoot, info.DefaultBranch)
	cmd = exec.Command("git", "-C", repoRoot, "worktree", "add", worktreePath, info.DefaultBranch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create default branch worktree: %w", err)
	}

	// Step 11: Copy untracked files from staging into worktree
	if len(untrackedFiles) > 0 {
		progress("Restoring untracked files")
		for _, file := range untrackedFiles {
			src := filepath.Join(stagingDir, file)
			dst := filepath.Join(worktreePath, file)

			// Create parent directories if needed
			if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for %s: %w", file, err)
			}

			// Move file (faster than copy)
			if err := os.Rename(src, dst); err != nil {
				// If rename fails (e.g., cross-device), fall back to copy
				if err := copyFile(src, dst); err != nil {
					return fmt.Errorf("failed to restore untracked file %s: %w", file, err)
				}
				_ = os.Remove(src)
			}
		}
	}

	// Step 12: Create feature branch worktree (if needed)
	if info.CurrentBranch != "" && info.CurrentBranch != info.DefaultBranch {
		progress(fmt.Sprintf("Creating worktree for %s", info.CurrentBranch))
		// Sanitize branch name: replace "/" with "-" for directory name
		sanitizedBranch := strings.ReplaceAll(info.CurrentBranch, "/", "-")
		featureWorktreePath := filepath.Join(repoRoot, sanitizedBranch)
		cmd = exec.Command("git", "-C", repoRoot, "worktree", "add", featureWorktreePath, info.CurrentBranch)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create feature branch worktree: %w", err)
		}
	}

	// Step 13: Clean up staging directory
	progress("Cleaning up staging area")
	if err := os.RemoveAll(stagingDir); err != nil {
		return fmt.Errorf("failed to remove staging directory: %w", err)
	}

	// Step 14: Create .pttconfig/default.yml
	progress("Creating .pttconfig")
	pttconfigDir := filepath.Join(repoRoot, ".pttconfig")
	if err := createDefaultYAMLConfig(pttconfigDir); err != nil {
		return err
	}

	// Success - clear cleanup stack
	cleanupStack = nil
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
