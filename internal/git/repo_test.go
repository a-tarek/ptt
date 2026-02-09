package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestWorktreePath_BareRepoOld(t *testing.T) {
	// Legacy test - replaced by TestWorktreePath_BareRepo below
	t.Skip("Replaced by comprehensive bare repo tests")
}

func TestWorktreePath_SiblingMode(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "wt-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock repo directory
	repoDir := filepath.Join(tmpDir, "myrepo")
	if err := os.Mkdir(repoDir, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	// Initialize a git repo
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Initialize git repo (required for IsBareRepository check)
	if err := execCmd("git", "init"); err != nil {
		t.Skip("git not available, skipping test")
	}

	// Test sibling mode path computation
	targetPath, err := WorktreePath(repoDir, "feature")
	if err != nil {
		t.Fatalf("WorktreePath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "myrepo-feature")
	if targetPath != expected {
		t.Errorf("expected %s, got %s", expected, targetPath)
	}
}

func TestWorktreePath_ExistingPath(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "wt-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock repo directory
	repoDir := filepath.Join(tmpDir, "myrepo")
	if err := os.Mkdir(repoDir, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	// Initialize git repo
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	if err := execCmd("git", "init"); err != nil {
		t.Skip("git not available, skipping test")
	}

	// Create the target directory to simulate conflict
	existingPath := filepath.Join(tmpDir, "myrepo-feature")
	if err := os.Mkdir(existingPath, 0755); err != nil {
		t.Fatalf("failed to create existing path: %v", err)
	}

	// Should return error for existing path
	_, err = WorktreePath(repoDir, "feature")
	if err == nil {
		t.Error("expected error for existing path, got nil")
	}
}

// Helper function to run git commands
func execCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// Helper function to create a ptt bare repo structure for testing
func createPttBareRepo(t *testing.T) (containerRoot string, cleanup func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "wt-test-bare-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cleanup = func() {
		os.RemoveAll(tmpDir)
	}

	// Create origin repo with initial commit
	originDir := filepath.Join(tmpDir, "origin")
	if err := os.Mkdir(originDir, 0755); err != nil {
		cleanup()
		t.Fatalf("failed to create origin dir: %v", err)
	}

	// Initialize origin repo
	if err := execCmd("git", "-C", originDir, "init"); err != nil {
		cleanup()
		t.Skip("git not available, skipping test")
	}

	// Configure git for testing (avoid warnings)
	execCmd("git", "-C", originDir, "config", "user.email", "test@example.com")
	execCmd("git", "-C", originDir, "config", "user.name", "Test User")

	// Create initial commit in origin
	testFile := filepath.Join(originDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test Repo\n"), 0644); err != nil {
		cleanup()
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := execCmd("git", "-C", originDir, "add", "README.md"); err != nil {
		cleanup()
		t.Fatalf("failed to add file: %v", err)
	}

	if err := execCmd("git", "-C", originDir, "commit", "-m", "Initial commit"); err != nil {
		cleanup()
		t.Fatalf("failed to commit: %v", err)
	}

	// Create container directory
	containerRoot = filepath.Join(tmpDir, "project-bare")
	if err := os.Mkdir(containerRoot, 0755); err != nil {
		cleanup()
		t.Fatalf("failed to create container dir: %v", err)
	}

	// Clone as bare repo into .bare
	bareDir := filepath.Join(containerRoot, ".bare")
	if err := execCmd("git", "clone", "--bare", originDir, bareDir); err != nil {
		cleanup()
		t.Fatalf("failed to clone bare repo: %v", err)
	}

	// Create .git file pointing to .bare
	gitFile := filepath.Join(containerRoot, ".git")
	if err := os.WriteFile(gitFile, []byte("gitdir: ./.bare\n"), 0644); err != nil {
		cleanup()
		t.Fatalf("failed to create .git file: %v", err)
	}

	// Configure fetch refspec
	if err := execCmd("git", "-C", containerRoot, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*"); err != nil {
		cleanup()
		t.Fatalf("failed to configure fetch: %v", err)
	}

	// Fetch from origin
	if err := execCmd("git", "-C", containerRoot, "fetch", "origin"); err != nil {
		cleanup()
		t.Fatalf("failed to fetch: %v", err)
	}

	// Create main worktree
	if err := execCmd("git", "-C", bareDir, "worktree", "add", "../main", "master"); err != nil {
		cleanup()
		t.Fatalf("failed to create main worktree: %v", err)
	}

	return containerRoot, cleanup
}

// TestBareRepoRoot_FromWorktree verifies BareRepoRoot() from inside a worktree
func TestBareRepoRoot_FromWorktree(t *testing.T) {
	containerRoot, cleanup := createPttBareRepo(t)
	defer cleanup()

	// Save original dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Change to worktree directory
	if err := os.Chdir(filepath.Join(containerRoot, "main")); err != nil {
		t.Fatalf("failed to chdir to worktree: %v", err)
	}

	// Call BareRepoRoot
	bareRoot, err := BareRepoRoot()
	if err != nil {
		t.Fatalf("BareRepoRoot() failed: %v", err)
	}

	// Resolve both to canonical paths for comparison (handles /var vs /private/var on macOS)
	expected, _ := filepath.EvalSymlinks(containerRoot)
	actual, _ := filepath.EvalSymlinks(bareRoot)

	// Should return container root
	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

// TestBareRepoRoot_FromContainerRoot verifies BareRepoRoot() from container root
func TestBareRepoRoot_FromContainerRoot(t *testing.T) {
	containerRoot, cleanup := createPttBareRepo(t)
	defer cleanup()

	// Save original dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Change to container root
	if err := os.Chdir(containerRoot); err != nil {
		t.Fatalf("failed to chdir to container root: %v", err)
	}

	// Call BareRepoRoot
	bareRoot, err := BareRepoRoot()
	if err != nil {
		t.Fatalf("BareRepoRoot() failed: %v", err)
	}

	// Resolve both to canonical paths for comparison
	expected, _ := filepath.EvalSymlinks(containerRoot)
	actual, _ := filepath.EvalSymlinks(bareRoot)

	// Should return container root
	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

// TestBareRepoRoot_FromSubdir verifies BareRepoRoot() from subdirectory of worktree
func TestBareRepoRoot_FromSubdir(t *testing.T) {
	containerRoot, cleanup := createPttBareRepo(t)
	defer cleanup()

	// Save original dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Create subdirectory in worktree
	mainWorktree := filepath.Join(containerRoot, "main")
	subdir := filepath.Join(mainWorktree, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Change to subdirectory
	if err := os.Chdir(subdir); err != nil {
		t.Fatalf("failed to chdir to subdir: %v", err)
	}

	// Call BareRepoRoot
	bareRoot, err := BareRepoRoot()
	if err != nil {
		t.Fatalf("BareRepoRoot() failed: %v", err)
	}

	// Resolve both to canonical paths for comparison
	expected, _ := filepath.EvalSymlinks(containerRoot)
	actual, _ := filepath.EvalSymlinks(bareRoot)

	// Should return container root
	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

// TestBareRepoRoot_NotBareRepo verifies BareRepoRoot() rejects regular repos
func TestBareRepoRoot_NotBareRepo(t *testing.T) {
	// Create a regular git repo
	tmpDir, err := os.MkdirTemp("", "wt-test-regular-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize regular repo
	if err := execCmd("git", "-C", tmpDir, "init"); err != nil {
		t.Skip("git not available, skipping test")
	}

	// Save original dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Change to regular repo
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Call BareRepoRoot - should return error
	_, err = BareRepoRoot()
	if err == nil {
		t.Error("expected error for regular repo, got nil")
	}
}

// TestWorktreePath_BareRepo verifies nested path computation in bare repo
func TestWorktreePath_BareRepo(t *testing.T) {
	containerRoot, cleanup := createPttBareRepo(t)
	defer cleanup()

	// Save original dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Change to container root
	if err := os.Chdir(containerRoot); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Call WorktreePath
	targetPath, err := WorktreePath(containerRoot, "feature")
	if err != nil {
		t.Fatalf("WorktreePath failed: %v", err)
	}

	// Resolve both to canonical paths for comparison
	expected, _ := filepath.EvalSymlinks(filepath.Join(containerRoot, "feature"))
	actual, _ := filepath.EvalSymlinks(targetPath)

	// Should return nested path
	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

// TestConfigRoot_BareRepo verifies ConfigRoot() in bare repo context
func TestConfigRoot_BareRepo(t *testing.T) {
	containerRoot, cleanup := createPttBareRepo(t)
	defer cleanup()

	// Save original dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Change to worktree
	mainWorktree := filepath.Join(containerRoot, "main")
	if err := os.Chdir(mainWorktree); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Call ConfigRoot
	configRoot, err := ConfigRoot()
	if err != nil {
		t.Fatalf("ConfigRoot() failed: %v", err)
	}

	// Resolve both to canonical paths for comparison
	expected, _ := filepath.EvalSymlinks(containerRoot)
	actual, _ := filepath.EvalSymlinks(configRoot)

	// Should return container root
	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

// TestConfigRoot_RegularRepo verifies ConfigRoot() in non-bare context
func TestConfigRoot_RegularRepo(t *testing.T) {
	// Create a regular git repo
	tmpDir, err := os.MkdirTemp("", "wt-test-regular-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize regular repo
	if err := execCmd("git", "-C", tmpDir, "init"); err != nil {
		t.Skip("git not available, skipping test")
	}

	// Configure git
	execCmd("git", "-C", tmpDir, "config", "user.email", "test@example.com")
	execCmd("git", "-C", tmpDir, "config", "user.name", "Test User")

	// Create initial commit
	testFile := filepath.Join(tmpDir, "README.md")
	os.WriteFile(testFile, []byte("test\n"), 0644)
	execCmd("git", "-C", tmpDir, "add", "README.md")
	execCmd("git", "-C", tmpDir, "commit", "-m", "Initial commit")

	// Save original dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Change to regular repo
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Call ConfigRoot
	configRoot, err := ConfigRoot()
	if err != nil {
		t.Fatalf("ConfigRoot() failed: %v", err)
	}

	// Should return home path (same as GetHomePath())
	homePath, err := GetHomePath()
	if err != nil {
		t.Fatalf("GetHomePath() failed: %v", err)
	}

	if configRoot != homePath {
		t.Errorf("expected %s, got %s", homePath, configRoot)
	}
}
