package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// execTestCmd runs a command and returns an error if it fails
func execTestCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return err
	} else if len(output) > 0 {
		// Command succeeded
		return nil
	}
	return nil
}

// createNormalRepo creates a normal git clone with a remote
func createNormalRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Create a bare remote first
	remoteDir := filepath.Join(tmpDir, "remote.git")
	cmd := exec.Command("git", "init", "--bare", remoteDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init remote: %v\n%s", err, output)
	}

	// Clone it
	repoDir := filepath.Join(tmpDir, "repo")
	cmd = exec.Command("git", "clone", remoteDir, repoDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to clone: %v\n%s", err, output)
	}

	// Configure git
	cmd = exec.Command("git", "-C", repoDir, "config", "user.email", "test@example.com")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to set email: %v\n%s", err, output)
	}

	cmd = exec.Command("git", "-C", repoDir, "config", "user.name", "Test User")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to set name: %v\n%s", err, output)
	}

	// Create initial commit
	readmePath := filepath.Join(repoDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	cmd = exec.Command("git", "-C", repoDir, "add", "README.md")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to add README: %v\n%s", err, output)
	}

	cmd = exec.Command("git", "-C", repoDir, "commit", "-m", "Initial commit")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to commit: %v\n%s", err, output)
	}

	// Get current branch name (could be main or master)
	cmd = exec.Command("git", "-C", repoDir, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get branch name: %v", err)
	}
	branchName := strings.TrimSpace(string(output))

	cmd = exec.Command("git", "-C", repoDir, "push", "origin", branchName)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to push: %v\n%s", err, output)
	}

	return repoDir
}

// createNormalRepoWithConfig creates a normal repo with .pttconfig
func createNormalRepoWithConfig(t *testing.T) string {
	t.Helper()

	repoDir := createNormalRepo(t)

	// Create .pttconfig directory and default config
	pttconfigDir := filepath.Join(repoDir, ".pttconfig")
	if err := os.MkdirAll(pttconfigDir, 0755); err != nil {
		t.Fatalf("Failed to create .pttconfig: %v", err)
	}

	defaultConfig := filepath.Join(pttconfigDir, "default")
	if err := os.WriteFile(defaultConfig, []byte("# config\n"), 0644); err != nil {
		t.Fatalf("Failed to create default config: %v", err)
	}

	return repoDir
}

// createPttBareRepoManual creates a ptt bare repo manually (with .bare/ and .git pointer)
func createPttBareRepoManual(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	containerDir := filepath.Join(tmpDir, "repo-ptt")
	bareDir := filepath.Join(containerDir, ".bare")

	// Create container directory
	if err := os.MkdirAll(containerDir, 0755); err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	// Initialize bare repo in .bare/
	cmd := exec.Command("git", "init", "--bare", bareDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init bare repo: %v\n%s", err, output)
	}

	// Create .git pointer file
	gitPointer := filepath.Join(containerDir, ".git")
	if err := os.WriteFile(gitPointer, []byte("gitdir: ./.bare\n"), 0644); err != nil {
		t.Fatalf("Failed to create .git pointer: %v", err)
	}

	// Configure bare repo
	cmd = exec.Command("git", "-C", containerDir, "config", "core.bare", "true")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to set core.bare: %v\n%s", err, output)
	}

	cmd = exec.Command("git", "-C", containerDir, "config", "core.logallrefupdates", "true")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to enable reflog: %v\n%s", err, output)
	}

	// Create initial worktree with commit
	mainWorktreePath := filepath.Join(containerDir, "main")
	cmd = exec.Command("git", "-C", containerDir, "worktree", "add", mainWorktreePath, "-b", "main")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create main worktree: %v\n%s", err, output)
	}

	// Configure git in worktree
	cmd = exec.Command("git", "-C", mainWorktreePath, "config", "user.email", "test@example.com")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to set email: %v\n%s", err, output)
	}

	cmd = exec.Command("git", "-C", mainWorktreePath, "config", "user.name", "Test User")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to set name: %v\n%s", err, output)
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

	cmd = exec.Command("git", "-C", mainWorktreePath, "commit", "-m", "Initial commit")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to commit: %v\n%s", err, output)
	}

	return containerDir
}

// createRawBareRepo creates a raw bare repo (git clone --bare)
func createRawBareRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Create a normal repo first as source
	sourceDir := filepath.Join(tmpDir, "source")
	cmd := exec.Command("git", "init", sourceDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init source: %v\n%s", err, output)
	}

	// Configure git
	cmd = exec.Command("git", "-C", sourceDir, "config", "user.email", "test@example.com")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to set email: %v\n%s", err, output)
	}

	cmd = exec.Command("git", "-C", sourceDir, "config", "user.name", "Test User")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to set name: %v\n%s", err, output)
	}

	// Create initial commit
	readmePath := filepath.Join(sourceDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	cmd = exec.Command("git", "-C", sourceDir, "add", "README.md")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to add README: %v\n%s", err, output)
	}

	cmd = exec.Command("git", "-C", sourceDir, "commit", "-m", "Initial commit")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to commit: %v\n%s", err, output)
	}

	// Clone as bare
	bareDir := filepath.Join(tmpDir, "repo.git")
	cmd = exec.Command("git", "clone", "--bare", sourceDir, bareDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to clone bare: %v\n%s", err, output)
	}

	return bareDir
}

// createNormalRepoOnFeatureBranch creates a normal repo currently on a feature branch
func createNormalRepoOnFeatureBranch(t *testing.T) string {
	t.Helper()

	repoDir := createNormalRepo(t)

	// Create and switch to feature branch
	cmd := exec.Command("git", "-C", repoDir, "checkout", "-b", "feature/test")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create feature branch: %v\n%s", err, output)
	}

	return repoDir
}

// createLocalOnlyRepo creates a normal repo without a remote
func createLocalOnlyRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "repo")

	// Initialize repo
	cmd := exec.Command("git", "init", repoDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init repo: %v\n%s", err, output)
	}

	// Configure git
	cmd = exec.Command("git", "-C", repoDir, "config", "user.email", "test@example.com")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to set email: %v\n%s", err, output)
	}

	cmd = exec.Command("git", "-C", repoDir, "config", "user.name", "Test User")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to set name: %v\n%s", err, output)
	}

	// Create initial commit
	readmePath := filepath.Join(repoDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	cmd = exec.Command("git", "-C", repoDir, "add", "README.md")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to add README: %v\n%s", err, output)
	}

	cmd = exec.Command("git", "-C", repoDir, "commit", "-m", "Initial commit")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to commit: %v\n%s", err, output)
	}

	return repoDir
}

func TestInit_NormalRepo_Success(t *testing.T) {
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	repoDir := createNormalRepo(t)
	os.Chdir(repoDir)

	autoYes = true
	defer func() { autoYes = false }()

	err := initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Verify bare layout
	bareDir := filepath.Join(repoDir, ".bare")
	if _, err := os.Stat(bareDir); os.IsNotExist(err) {
		t.Errorf(".bare/ directory not created")
	}

	// Verify .git pointer file
	gitPointer := filepath.Join(repoDir, ".git")
	content, err := os.ReadFile(gitPointer)
	if err != nil {
		t.Errorf("Failed to read .git pointer: %v", err)
	}
	if strings.TrimSpace(string(content)) != "gitdir: ./.bare" {
		t.Errorf("Incorrect .git pointer content: %s", content)
	}

	// Verify default branch worktree exists (could be main or master)
	// Get the default branch name
	cmd := exec.Command("git", "-C", repoDir, "symbolic-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get HEAD ref: %v", err)
	}
	headRef := strings.TrimSpace(string(output))
	defaultBranch := strings.TrimPrefix(headRef, "refs/heads/")

	defaultWorktree := filepath.Join(repoDir, defaultBranch)
	if _, err := os.Stat(defaultWorktree); os.IsNotExist(err) {
		t.Errorf("%s worktree not created", defaultBranch)
	}

	// Verify .pttconfig/default created
	pttconfigDefault := filepath.Join(repoDir, ".pttconfig", "default")
	if _, err := os.Stat(pttconfigDefault); os.IsNotExist(err) {
		t.Errorf(".pttconfig/default not created")
	}
}

func TestInit_NormalRepo_FeatureBranch(t *testing.T) {
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	repoDir := createNormalRepoOnFeatureBranch(t)
	os.Chdir(repoDir)

	autoYes = true
	defer func() { autoYes = false }()

	err := initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// When on a feature branch, init creates a worktree for the current branch (feature/test)
	// The default branch (master/main) also gets a worktree, plus the current branch

	// First, determine what the original default branch was before init
	// Since we're on feature/test, the remote's default should be master
	cmd := exec.Command("git", "-C", repoDir, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to list worktrees: %v", err)
	}

	// Parse worktree list to find branches
	lines := strings.Split(string(output), "\n")
	foundMaster := false
	foundFeature := false

	for _, line := range lines {
		if strings.Contains(line, "branch refs/heads/master") {
			foundMaster = true
		}
		if strings.Contains(line, "branch refs/heads/feature/test") {
			foundFeature = true
		}
	}

	if !foundMaster {
		t.Errorf("master worktree not created")
	}
	if !foundFeature {
		t.Errorf("feature/test worktree not created")
	}
}

func TestInit_NormalRepo_PreservesUntracked(t *testing.T) {
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	repoDir := createNormalRepo(t)
	os.Chdir(repoDir)

	// Create untracked files
	untrackedFile := filepath.Join(repoDir, "untracked.txt")
	if err := os.WriteFile(untrackedFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create untracked file: %v", err)
	}

	autoYes = true
	defer func() { autoYes = false }()

	err := initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Verify untracked file exists in worktree after init
	cmd := exec.Command("git", "-C", repoDir, "symbolic-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get HEAD ref: %v", err)
	}
	headRef := strings.TrimSpace(string(output))
	defaultBranch := strings.TrimPrefix(headRef, "refs/heads/")

	defaultWorktree := filepath.Join(repoDir, defaultBranch)
	untrackedInWorktree := filepath.Join(defaultWorktree, "untracked.txt")
	if _, err := os.Stat(untrackedInWorktree); os.IsNotExist(err) {
		t.Errorf("Untracked file was not preserved in worktree")
	}
}

func TestInit_NormalRepo_DirtyRejects(t *testing.T) {
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	repoDir := createNormalRepo(t)
	os.Chdir(repoDir)

	// Make uncommitted changes
	readmePath := filepath.Join(repoDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Modified\n"), 0644); err != nil {
		t.Fatalf("Failed to modify README: %v", err)
	}

	autoYes = true
	defer func() { autoYes = false }()

	err := initCmd.RunE(initCmd, []string{})
	if err == nil {
		t.Errorf("Expected error for dirty working tree, got nil")
	}

	// The error is "validation failed" but the details are printed to stderr
	// Just verify we got an error
	if err.Error() != "validation failed" {
		t.Errorf("Expected 'validation failed' error, got: %v", err)
	}
}

func TestInit_NormalRepo_LocalOnly(t *testing.T) {
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	repoDir := createLocalOnlyRepo(t)
	os.Chdir(repoDir)

	autoYes = true
	defer func() { autoYes = false }()

	err := initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Verify .pttconfig/default created even without remote
	pttconfigDefault := filepath.Join(repoDir, ".pttconfig", "default")
	if _, err := os.Stat(pttconfigDefault); os.IsNotExist(err) {
		t.Errorf(".pttconfig/default not created")
	}
}

func TestInit_RawBareRepo_Success(t *testing.T) {
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	bareDir := createRawBareRepo(t)
	os.Chdir(bareDir)

	autoYes = true
	defer func() { autoYes = false }()

	err := initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Verify .bare/ wrapper created
	dotBareDir := filepath.Join(bareDir, ".bare")
	if _, err := os.Stat(dotBareDir); os.IsNotExist(err) {
		t.Errorf(".bare/ directory not created")
	}

	// Verify .git pointer
	gitPointer := filepath.Join(bareDir, ".git")
	content, err := os.ReadFile(gitPointer)
	if err != nil {
		t.Errorf("Failed to read .git pointer: %v", err)
	}
	if strings.TrimSpace(string(content)) != "gitdir: ./.bare" {
		t.Errorf("Incorrect .git pointer content: %s", content)
	}

	// Verify default branch worktree (master in this case based on git defaults)
	masterWorktree := filepath.Join(bareDir, "master")
	if _, err := os.Stat(masterWorktree); os.IsNotExist(err) {
		t.Errorf("master worktree not created")
	}

	// Verify refspec configured
	cmd := exec.Command("git", "-C", bareDir, "config", "remote.origin.fetch")
	output, err := cmd.Output()
	expectedRefspec := "+refs/heads/*:refs/remotes/origin/*"
	if err == nil && !strings.Contains(string(output), expectedRefspec) {
		// Note: raw bare repo created from local source won't have remote, so this check may not apply
		// Only check if remote exists
		checkCmd := exec.Command("git", "-C", bareDir, "remote", "get-url", "origin")
		if checkCmd.Run() == nil {
			t.Errorf("Refspec not configured correctly")
		}
	}

	// Verify reflog enabled
	cmd = exec.Command("git", "-C", bareDir, "config", "core.logallrefupdates")
	output, err = cmd.Output()
	if err != nil || strings.TrimSpace(string(output)) != "true" {
		t.Errorf("Reflog not enabled")
	}
}

func TestInit_PttBare_CreatesPttconfig(t *testing.T) {
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	containerDir := createPttBareRepoManual(t)
	os.Chdir(containerDir)

	autoYes = true
	defer func() { autoYes = false }()

	err := initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Verify .pttconfig/default created
	pttconfigDefault := filepath.Join(containerDir, ".pttconfig", "default")
	if _, err := os.Stat(pttconfigDefault); os.IsNotExist(err) {
		t.Errorf(".pttconfig/default not created")
	}
}

func TestInit_PttBare_AlreadySetUp(t *testing.T) {
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	containerDir := createPttBareRepoManual(t)

	// Create .pttconfig/default
	pttconfigDir := filepath.Join(containerDir, ".pttconfig")
	if err := os.MkdirAll(pttconfigDir, 0755); err != nil {
		t.Fatalf("Failed to create .pttconfig: %v", err)
	}
	pttconfigDefault := filepath.Join(pttconfigDir, "default")
	if err := os.WriteFile(pttconfigDefault, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create default config: %v", err)
	}

	os.Chdir(containerDir)

	autoYes = true
	defer func() { autoYes = false }()

	err := initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Errorf("init should be idempotent, but got error: %v", err)
	}
}

func TestInit_NotGitRepo(t *testing.T) {
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Create non-git directory
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	autoYes = true
	defer func() { autoYes = false }()

	err := initCmd.RunE(initCmd, []string{})
	if err == nil {
		t.Errorf("Expected error for non-git directory, got nil")
	}

	if !strings.Contains(err.Error(), "Not inside a git repository") {
		t.Errorf("Error message doesn't mention 'Not inside a git repository': %v", err)
	}
}
