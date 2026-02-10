package initcmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AdoptRawBareRepo converts a raw bare repository into a ptt-managed bare repo in place.
// For raw bare repos, the git data IS the repoRoot itself (HEAD, objects/, refs/, etc.).
//
// Algorithm:
// 1. Create .bare/ directory inside repoRoot
// 2. Move all git data entries into .bare/
// 3. Create .git pointer file
// 4. Configure fetch refspec (if has remote)
// 5. Enable reflog
// 6. Fetch latest (if has remote, best-effort)
// 7. Move external worktrees inside container (skip if already inside)
// 8. Repair worktree pointers
// 9. Create default branch worktree if not already present
// 10. Create .pttconfig/default
func AdoptRawBareRepo(repoRoot string, info *RepoInfo, progress ProgressCallback) error {
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

	bareDir := filepath.Join(repoRoot, ".bare")

	// Step 1: Create .bare/ directory
	progress("Creating .bare directory")
	if err := os.MkdirAll(bareDir, 0755); err != nil {
		return fmt.Errorf("failed to create .bare directory: %w", err)
	}
	cleanupStack = append(cleanupStack, func() error {
		return os.RemoveAll(bareDir)
	})

	// Step 2: Move all git data entries into .bare/
	progress("Moving git data into .bare")

	entries, err := os.ReadDir(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to read repo root: %w", err)
	}

	var movedEntries []string
	for _, entry := range entries {
		name := entry.Name()

		// Skip the .bare directory we just created
		if name == ".bare" {
			continue
		}

		// Move everything (bare repos should only contain git data)
		src := filepath.Join(repoRoot, name)
		dst := filepath.Join(bareDir, name)

		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("failed to move %s to .bare: %w", name, err)
		}
		movedEntries = append(movedEntries, name)
	}

	// Register cleanup to move everything back
	cleanupStack = append(cleanupStack, func() error {
		for _, name := range movedEntries {
			src := filepath.Join(bareDir, name)
			dst := filepath.Join(repoRoot, name)
			_ = os.Rename(src, dst)
		}
		return nil
	})

	// Step 3: Create .git pointer file
	progress("Creating .git pointer file")
	gitPointer := filepath.Join(repoRoot, ".git")
	if err := os.WriteFile(gitPointer, []byte("gitdir: ./.bare\n"), 0644); err != nil {
		return fmt.Errorf("failed to create .git pointer file: %w", err)
	}

	// Step 4: Configure fetch refspec (if has remote)
	if info.HasRemote {
		progress("Configuring fetch refspec")
		cmd := exec.Command("git", "-C", repoRoot, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to configure fetch refspec: %w", err)
		}
	}

	// Step 5: Enable reflog
	progress("Enabling reflog")
	cmd := exec.Command("git", "-C", repoRoot, "config", "core.logallrefupdates", "true")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable reflog: %w", err)
	}

	// Step 6: Fetch latest (if has remote, best-effort)
	if info.HasRemote {
		progress("Fetching from remote")
		cmd = exec.Command("git", "-C", repoRoot, "fetch", "origin")
		// Ignore errors - network issues shouldn't block adoption
		_ = cmd.Run()
	}

	// Step 7: Move external worktrees inside container
	if len(info.ExistingWorktrees) > 0 {
		progress("Checking worktree locations")
		for _, wt := range info.ExistingWorktrees {
			// Skip if worktree is already inside the container
			relPath, err := filepath.Rel(repoRoot, wt.Path)
			if err == nil && !strings.HasPrefix(relPath, "..") {
				// Already inside container
				continue
			}

			// Move external worktree inside
			wtName := filepath.Base(wt.Path)
			newPath := filepath.Join(repoRoot, wtName)

			// Check if target already exists
			if _, err := os.Stat(newPath); err == nil {
				// Try with branch name as suffix
				if wt.Branch != "" {
					sanitizedBranch := strings.ReplaceAll(wt.Branch, "/", "-")
					newPath = filepath.Join(repoRoot, sanitizedBranch)
				}
			}

			progress(fmt.Sprintf("Moving worktree %s inside container", wtName))
			if err := os.Rename(wt.Path, newPath); err != nil {
				// Non-fatal: log but continue
				progress(fmt.Sprintf("Warning: could not move worktree %s: %v", wtName, err))
			}
		}
	}

	// Step 8: Repair worktree pointers
	if len(info.ExistingWorktrees) > 0 {
		progress("Repairing worktree pointers")
		cmd = exec.Command("git", "-C", repoRoot, "worktree", "repair")
		// Best-effort - don't fail if repair has issues
		_ = cmd.Run()
	}

	// Step 9: Create default branch worktree if not already present
	hasDefaultWorktree := false
	for _, wt := range info.ExistingWorktrees {
		if wt.Branch == info.DefaultBranch {
			hasDefaultWorktree = true
			break
		}
	}

	if !hasDefaultWorktree && info.DefaultBranch != "" {
		progress(fmt.Sprintf("Creating worktree for %s", info.DefaultBranch))
		worktreePath := filepath.Join(repoRoot, info.DefaultBranch)
		cmd = exec.Command("git", "-C", repoRoot, "worktree", "add", worktreePath, info.DefaultBranch)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create default branch worktree: %w", err)
		}
	}

	// Step 10: Create .pttconfig/default.yml
	progress("Creating .pttconfig")
	pttconfigDir := filepath.Join(repoRoot, ".pttconfig")
	if err := createDefaultYAMLConfig(pttconfigDir); err != nil {
		return err
	}

	// Success - clear cleanup stack
	cleanupStack = nil
	return nil
}
