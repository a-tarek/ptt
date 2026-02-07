package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupBareRepoWithCommit creates a bare repo with a main worktree that has at least one commit
func setupBareRepoWithCommit(t *testing.T) (bareRoot string, mainWorktreePath string) {
	t.Helper()

	// Create temporary directory for bare repo
	bareRoot = t.TempDir()
	bareGitDir := filepath.Join(bareRoot, ".git")

	// Initialize bare repo
	cmd := exec.Command("git", "init", "--bare", bareGitDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init bare repo: %v", err)
	}

	// Create main worktree (following wt naming convention: bareRoot/branchname)
	mainWorktreePath = filepath.Join(bareRoot, "main")
	cmd = exec.Command("git", "-C", bareGitDir, "worktree", "add", mainWorktreePath, "-b", "main")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create main worktree: %v", err)
	}

	// Configure git user (required for commits)
	cmd = exec.Command("git", "-C", mainWorktreePath, "config", "user.name", "Test User")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to set git user.name: %v", err)
	}
	cmd = exec.Command("git", "-C", mainWorktreePath, "config", "user.email", "test@example.com")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to set git user.email: %v", err)
	}

	// Create initial commit
	readmePath := filepath.Join(mainWorktreePath, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test repo\n"), 0644); err != nil {
		t.Fatalf("failed to create README: %v", err)
	}
	cmd = exec.Command("git", "-C", mainWorktreePath, "add", "README.md")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to git add: %v", err)
	}
	cmd = exec.Command("git", "-C", mainWorktreePath, "commit", "-m", "Initial commit")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	return bareRoot, mainWorktreePath
}

func TestEjectFromHomeWorktree(t *testing.T) {
	bareRoot, mainWorktreePath := setupBareRepoWithCommit(t)

	// Change to main worktree
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(mainWorktreePath); err != nil {
		t.Fatalf("failed to chdir to main worktree: %v", err)
	}

	// Create a feature branch
	cmd := exec.Command("git", "checkout", "-b", "feature-x")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create feature branch: %v", err)
	}

	// Create a dirty file (uncommitted change)
	dirtyPath := filepath.Join(mainWorktreePath, "dirty.txt")
	if err := os.WriteFile(dirtyPath, []byte("uncommitted\n"), 0644); err != nil {
		t.Fatalf("failed to create dirty file: %v", err)
	}

	// Run wt eject
	rootCmd.SetArgs([]string{"eject"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("wt eject failed: %v", err)
	}

	// Reset args
	rootCmd.SetArgs([]string{})

	// Verify: new worktree exists at {bareRoot}/feature-x
	expectedPath := filepath.Join(bareRoot, "feature-x")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("new worktree not created at %s", expectedPath)
	}

	// Verify: main worktree is now on "main" branch
	cmd = exec.Command("git", "-C", mainWorktreePath, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to get current branch: %v", err)
	}
	currentBranch := strings.TrimSpace(string(output))
	if currentBranch != "main" {
		t.Errorf("main worktree should be on 'main', got '%s'", currentBranch)
	}

	// Verify: dirty file exists in new worktree (stash was popped)
	dirtyInNewWT := filepath.Join(expectedPath, "dirty.txt")
	if _, err := os.Stat(dirtyInNewWT); os.IsNotExist(err) {
		t.Errorf("dirty file not found in new worktree (stash was not popped)")
	}
}

func TestEjectWithCustomName(t *testing.T) {
	bareRoot, mainWorktreePath := setupBareRepoWithCommit(t)

	// Change to main worktree
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(mainWorktreePath); err != nil {
		t.Fatalf("failed to chdir to main worktree: %v", err)
	}

	// Create a feature branch
	cmd := exec.Command("git", "checkout", "-b", "feature-x")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create feature branch: %v", err)
	}

	// Run wt eject with custom name
	rootCmd.SetArgs([]string{"eject", "my-feature"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("wt eject failed: %v", err)
	}

	// Reset args
	rootCmd.SetArgs([]string{})

	// Verify: worktree created at {bareRoot}/my-feature (not feature-x)
	expectedPath := filepath.Join(bareRoot, "my-feature")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("new worktree not created at %s (should use custom name)", expectedPath)
	}

	// Verify feature-x path does NOT exist
	wrongPath := filepath.Join(bareRoot, "feature-x")
	if _, err := os.Stat(wrongPath); err == nil {
		t.Errorf("worktree created at %s (should be my-feature)", wrongPath)
	}
}

func TestEjectDetachedHeadError(t *testing.T) {
	_, mainWorktreePath := setupBareRepoWithCommit(t)

	// Change to main worktree
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(mainWorktreePath); err != nil {
		t.Fatalf("failed to chdir to main worktree: %v", err)
	}

	// Detach HEAD
	cmd := exec.Command("git", "checkout", "--detach")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to detach HEAD: %v", err)
	}

	// Run wt eject
	rootCmd.SetArgs([]string{"eject"})
	err = rootCmd.Execute()

	// Reset args
	rootCmd.SetArgs([]string{})

	// Verify: returns error containing "detached HEAD"
	if err == nil {
		t.Fatal("wt eject should fail on detached HEAD")
	}
	if !strings.Contains(err.Error(), "detached HEAD") {
		t.Errorf("error should mention 'detached HEAD', got: %v", err)
	}
}

func TestEjectAlreadyOnFallbackError(t *testing.T) {
	_, mainWorktreePath := setupBareRepoWithCommit(t)

	// Change to main worktree
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(mainWorktreePath); err != nil {
		t.Fatalf("failed to chdir to main worktree: %v", err)
	}

	// We're on "main" branch (the fallback for home worktree)
	// Run wt eject
	rootCmd.SetArgs([]string{"eject"})
	err = rootCmd.Execute()

	// Reset args
	rootCmd.SetArgs([]string{})

	// Verify: returns error containing "already on 'main'"
	if err == nil {
		t.Fatal("wt eject should fail when already on fallback branch")
	}
	if !strings.Contains(err.Error(), "already on 'main'") {
		t.Errorf("error should mention 'already on main', got: %v", err)
	}
}
