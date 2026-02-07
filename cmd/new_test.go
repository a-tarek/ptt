package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupBareRepo creates a temporary bare git repository with an initial worktree
func setupBareRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	bareRepoPath := filepath.Join(tmpDir, "repo.git")

	// Initialize bare repository
	cmd := exec.Command("git", "init", "--bare", bareRepoPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init bare repo: %v\n%s", err, output)
	}

	// Create initial worktree with a commit
	mainWorktreePath := filepath.Join(bareRepoPath, "main")
	cmd = exec.Command("git", "-C", bareRepoPath, "worktree", "add", mainWorktreePath, "-b", "main")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create main worktree: %v\n%s", err, output)
	}

	// Create a file and commit
	testFile := filepath.Join(mainWorktreePath, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	cmd = exec.Command("git", "-C", mainWorktreePath, "add", "README.md")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to add README: %v\n%s", err, output)
	}

	cmd = exec.Command("git", "-C", mainWorktreePath, "config", "user.email", "test@example.com")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to set email: %v\n%s", err, output)
	}

	cmd = exec.Command("git", "-C", mainWorktreePath, "config", "user.name", "Test User")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to set name: %v\n%s", err, output)
	}

	cmd = exec.Command("git", "-C", mainWorktreePath, "commit", "-m", "Initial commit")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to commit: %v\n%s", err, output)
	}

	return bareRepoPath
}

func TestNewCommandCreatesWorktree(t *testing.T) {
	bareRoot := setupBareRepo(t)
	mainWorktreePath := filepath.Join(bareRoot, "main")

	// Save original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Change to main worktree
	if err := os.Chdir(mainWorktreePath); err != nil {
		t.Fatalf("Failed to change to main worktree: %v", err)
	}

	// Execute wt new staging
	rootCmd.SetArgs([]string{"new", "staging"})
	defer rootCmd.SetArgs([]string{}) // Reset args

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("wt new failed: %v", err)
	}

	// Verify worktree directory exists
	stagingPath := filepath.Join(bareRoot, "staging")
	if _, err := os.Stat(stagingPath); os.IsNotExist(err) {
		t.Errorf("Worktree directory not created at %s", stagingPath)
	}

	// Verify branch was created
	cmd := exec.Command("git", "-C", mainWorktreePath, "branch", "--list", "staging")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to list branches: %v", err)
	}

	if len(output) == 0 {
		t.Errorf("Branch 'staging' was not created")
	}
}

func TestNewCommandExistingBranch(t *testing.T) {
	bareRoot := setupBareRepo(t)
	mainWorktreePath := filepath.Join(bareRoot, "main")

	// Save original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Change to main worktree
	if err := os.Chdir(mainWorktreePath); err != nil {
		t.Fatalf("Failed to change to main worktree: %v", err)
	}

	// Create a branch 'feature' first
	cmd := exec.Command("git", "-C", mainWorktreePath, "branch", "feature")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create feature branch: %v\n%s", err, output)
	}

	// Execute wt new feature (should use existing branch)
	rootCmd.SetArgs([]string{"new", "feature"})
	defer rootCmd.SetArgs([]string{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("wt new failed with existing branch: %v", err)
	}

	// Verify worktree directory exists
	featurePath := filepath.Join(bareRoot, "feature")
	if _, err := os.Stat(featurePath); os.IsNotExist(err) {
		t.Errorf("Worktree directory not created at %s", featurePath)
	}
}

func TestNewCommandWithSkipConfig(t *testing.T) {
	bareRoot := setupBareRepo(t)
	mainWorktreePath := filepath.Join(bareRoot, "main")

	// Save original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create .wtconfig with copy action
	wtconfigPath := filepath.Join(bareRoot, ".wtconfig")
	if err := os.WriteFile(wtconfigPath, []byte("copy .env\n"), 0644); err != nil {
		t.Fatalf("Failed to create .wtconfig: %v", err)
	}

	// Create .env file in main worktree
	envFile := filepath.Join(mainWorktreePath, ".env")
	if err := os.WriteFile(envFile, []byte("TEST=1\n"), 0644); err != nil {
		t.Fatalf("Failed to create .env: %v", err)
	}

	// Change to main worktree
	if err := os.Chdir(mainWorktreePath); err != nil {
		t.Fatalf("Failed to change to main worktree: %v", err)
	}

	// Execute wt new --skip-config staging
	rootCmd.SetArgs([]string{"new", "--skip-config", "staging"})
	defer rootCmd.SetArgs([]string{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("wt new failed: %v", err)
	}

	// Verify .env was NOT copied
	stagingPath := filepath.Join(bareRoot, "staging")
	stagingEnvFile := filepath.Join(stagingPath, ".env")
	if _, err := os.Stat(stagingEnvFile); err == nil {
		t.Errorf(".env should not have been copied with --skip-config")
	}
}

func TestNewCommandAlreadyExists(t *testing.T) {
	bareRoot := setupBareRepo(t)
	mainWorktreePath := filepath.Join(bareRoot, "main")

	// Save original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create directory at target path
	stagingPath := filepath.Join(bareRoot, "staging")
	if err := os.MkdirAll(stagingPath, 0755); err != nil {
		t.Fatalf("Failed to create staging directory: %v", err)
	}

	// Change to main worktree
	if err := os.Chdir(mainWorktreePath); err != nil {
		t.Fatalf("Failed to change to main worktree: %v", err)
	}

	// Execute wt new staging (should fail)
	rootCmd.SetArgs([]string{"new", "staging"})
	defer rootCmd.SetArgs([]string{})

	err = rootCmd.Execute()
	if err == nil {
		t.Errorf("Expected error when path already exists, got nil")
	}
}
