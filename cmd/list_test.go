package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
)

// setupPttBareRepo creates a ptt-style bare repository with a worktree
func setupPttBareRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Create origin repo with initial commit
	originDir := filepath.Join(tmpDir, "origin")
	if err := os.Mkdir(originDir, 0755); err != nil {
		t.Fatalf("failed to create origin dir: %v", err)
	}

	// Initialize origin repo
	cmd := exec.Command("git", "-C", originDir, "init")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("git not available, skipping test: %v\n%s", err, output)
	}

	// Configure git for testing
	exec.Command("git", "-C", originDir, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", originDir, "config", "user.name", "Test User").Run()

	// Create initial commit in origin
	testFile := filepath.Join(originDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test Repo\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "-C", originDir, "add", "README.md")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to add file: %v\n%s", err, output)
	}

	cmd = exec.Command("git", "-C", originDir, "commit", "-m", "Initial commit")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to commit: %v\n%s", err, output)
	}

	// Create container directory
	containerRoot := filepath.Join(tmpDir, "project-bare")
	if err := os.Mkdir(containerRoot, 0755); err != nil {
		t.Fatalf("failed to create container dir: %v", err)
	}

	// Clone as bare repo into .bare
	bareDir := filepath.Join(containerRoot, ".bare")
	cmd = exec.Command("git", "clone", "--bare", originDir, bareDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to clone bare repo: %v\n%s", err, output)
	}

	// Create .git file pointing to .bare
	gitFile := filepath.Join(containerRoot, ".git")
	if err := os.WriteFile(gitFile, []byte("gitdir: ./.bare\n"), 0644); err != nil {
		t.Fatalf("failed to create .git file: %v", err)
	}

	// Configure fetch refspec
	cmd = exec.Command("git", "-C", containerRoot, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to configure fetch: %v\n%s", err, output)
	}

	// Fetch from origin
	cmd = exec.Command("git", "-C", containerRoot, "fetch", "origin")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to fetch: %v\n%s", err, output)
	}

	// Create main worktree
	cmd = exec.Command("git", "-C", bareDir, "worktree", "add", "../main", "master")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create main worktree: %v\n%s", err, output)
	}

	return containerRoot
}

// setupRegularRepo creates a standard git repository
func setupRegularRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Initialize regular repo
	cmd := exec.Command("git", "-C", tmpDir, "init")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("git not available, skipping test: %v\n%s", err, output)
	}

	// Configure git
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Create initial commit
	testFile := filepath.Join(tmpDir, "README.md")
	os.WriteFile(testFile, []byte("test\n"), 0644)
	exec.Command("git", "-C", tmpDir, "add", "README.md").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "Initial commit").Run()

	return tmpDir
}

func TestListCmd_FiltersBareEntry(t *testing.T) {
	containerRoot := setupPttBareRepo(t)

	// Save original dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Change to main worktree
	mainWorktree := filepath.Join(containerRoot, "main")
	if err := os.Chdir(mainWorktree); err != nil {
		t.Fatalf("failed to chdir to worktree: %v", err)
	}

	// First verify that git shows both bare and regular worktrees
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	porcelainOutput, err := cmd.Output()
	if err != nil {
		t.Fatalf("git worktree list failed: %v", err)
	}

	gitOutput := string(porcelainOutput)
	t.Logf("Git worktree list porcelain:\n%s", gitOutput)

	// Verify git shows both .bare and main worktrees
	if !strings.Contains(gitOutput, ".bare") {
		t.Skip("git doesn't report .bare entry, test environment issue")
	}
	if !strings.Contains(gitOutput, "/main") {
		t.Skip("git doesn't report main worktree, test environment issue")
	}
	if !strings.Contains(gitOutput, "bare") {
		t.Skip("git doesn't mark .bare as bare, test environment issue")
	}

	// Disable color for test output
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = originalNoColor }()

	// Capture stdout by redirecting it temporarily
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute command
	rootCmd.SetArgs([]string{"ls"})
	defer rootCmd.SetArgs([]string{})

	execErr := rootCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if execErr != nil {
		t.Fatalf("ls command failed: %v", execErr)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	t.Logf("ptt ls output:\n%s", output)

	// KEY TEST: Verify .bare is NOT in ptt ls output
	// (even though git reports it)
	if strings.Contains(output, ".bare") {
		t.Errorf("output should not contain .bare metadata entry, got:\n%s", output)
	}

	// Verify output is not empty (should show main worktree)
	if len(strings.TrimSpace(output)) == 0 {
		t.Errorf("output should not be empty, expected to show main worktree")
	}

	// Verify master branch appears (the worktree we expect to see)
	if !strings.Contains(output, "master") {
		t.Errorf("output should contain master branch, got:\n%s", output)
	}
}

func TestListCmd_NormalRepo(t *testing.T) {
	repoDir := setupRegularRepo(t)

	// Save original dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Change to repo directory
	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Disable color for test output
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = originalNoColor }()

	// Capture stdout by redirecting it temporarily
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute command
	rootCmd.SetArgs([]string{"ls"})
	defer rootCmd.SetArgs([]string{})

	err := rootCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Should work without error (normal repos show repo root as worktree)
	if err != nil {
		t.Fatalf("ls command failed in normal repo: %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// In a normal repo with no additional worktrees, git worktree list
	// shows the repo itself. This is expected behavior and should work unchanged.
	// We just verify no error occurred.
	t.Logf("Normal repo output: %s", output)
}
