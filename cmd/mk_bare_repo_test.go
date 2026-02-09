package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// execTestCmd runs a command for testing
func execTestCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// createNormalRepo creates a normal git repo with a remote origin
func createNormalRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "repo")
	remoteDir := filepath.Join(tmpDir, "remote.git")

	// Create repo
	if err := os.Mkdir(repoDir, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	if err := execTestCmd("git", "-C", repoDir, "init", "-b", "main"); err != nil {
		t.Skip("git not available, skipping test")
	}

	// Configure git
	execTestCmd("git", "-C", repoDir, "config", "user.email", "test@example.com")
	execTestCmd("git", "-C", repoDir, "config", "user.name", "Test User")

	// Create initial commit
	if err := execTestCmd("git", "-C", repoDir, "commit", "--allow-empty", "-m", "init"); err != nil {
		t.Fatalf("failed to create initial commit: %v", err)
	}

	// Create remote
	if err := execTestCmd("git", "init", "--bare", remoteDir); err != nil {
		t.Fatalf("failed to create remote: %v", err)
	}

	// Add remote and push
	if err := execTestCmd("git", "-C", repoDir, "remote", "add", "origin", remoteDir); err != nil {
		t.Fatalf("failed to add remote: %v", err)
	}

	if err := execTestCmd("git", "-C", repoDir, "push", "-u", "origin", "main"); err != nil {
		t.Fatalf("failed to push to remote: %v", err)
	}

	return repoDir
}

// createNormalRepoWithConfig creates a normal repo with .pttconfig
func createNormalRepoWithConfig(t *testing.T) string {
	t.Helper()

	repoDir := createNormalRepo(t)

	// Create .pttconfig/default
	pttConfigDir := filepath.Join(repoDir, ".pttconfig")
	if err := os.Mkdir(pttConfigDir, 0755); err != nil {
		t.Fatalf("failed to create .pttconfig: %v", err)
	}

	defaultConfig := filepath.Join(pttConfigDir, "default")
	if err := os.WriteFile(defaultConfig, []byte("copy .env\n"), 0644); err != nil {
		t.Fatalf("failed to write default config: %v", err)
	}

	return repoDir
}

// createPttBareRepoManual creates a ptt bare repo manually for testing
func createPttBareRepoManual(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	containerRoot := filepath.Join(tmpDir, "project-bare")

	// Create remote first
	remoteDir := filepath.Join(tmpDir, "remote.git")
	if err := execTestCmd("git", "init", "--bare", remoteDir); err != nil {
		t.Fatalf("failed to create remote: %v", err)
	}

	// Create a temp repo to push initial commit
	tempRepo := filepath.Join(tmpDir, "temp")
	if err := os.Mkdir(tempRepo, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	execTestCmd("git", "-C", tempRepo, "init", "-b", "main")
	execTestCmd("git", "-C", tempRepo, "config", "user.email", "test@example.com")
	execTestCmd("git", "-C", tempRepo, "config", "user.name", "Test User")
	execTestCmd("git", "-C", tempRepo, "commit", "--allow-empty", "-m", "init")
	execTestCmd("git", "-C", tempRepo, "remote", "add", "origin", remoteDir)
	execTestCmd("git", "-C", tempRepo, "push", "-u", "origin", "main")

	// Create ptt bare repo
	if err := os.Mkdir(containerRoot, 0755); err != nil {
		t.Fatalf("failed to create container: %v", err)
	}

	bareDir := filepath.Join(containerRoot, ".bare")
	execTestCmd("git", "clone", "--bare", remoteDir, bareDir)
	os.WriteFile(filepath.Join(containerRoot, ".git"), []byte("gitdir: ./.bare\n"), 0644)
	execTestCmd("git", "-C", containerRoot, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
	execTestCmd("git", "-C", containerRoot, "config", "core.logallrefupdates", "true")
	execTestCmd("git", "-C", containerRoot, "fetch", "origin")
	execTestCmd("git", "-C", bareDir, "worktree", "add", filepath.Join(containerRoot, "main"), "main")

	return containerRoot
}

func TestMkBareRepo_Success(t *testing.T) {
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	repoDir := createNormalRepo(t)
	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Execute mk-bare-repo
	if err := mkBareRepoCmd.RunE(mkBareRepoCmd, []string{}); err != nil {
		t.Fatalf("mk-bare-repo failed: %v", err)
	}

	// Verify target directory exists
	targetDir := filepath.Join(filepath.Dir(repoDir), "repo-bare")
	if _, err := os.Stat(targetDir); err != nil {
		t.Errorf("target directory not created: %v", err)
	}

	// Verify .bare directory exists
	bareDir := filepath.Join(targetDir, ".bare")
	if info, err := os.Stat(bareDir); err != nil || !info.IsDir() {
		t.Errorf(".bare directory not found")
	}

	// Verify .git pointer file
	gitFile := filepath.Join(targetDir, ".git")
	content, err := os.ReadFile(gitFile)
	if err != nil {
		t.Errorf("failed to read .git file: %v", err)
	}
	if string(content) != "gitdir: ./.bare\n" {
		t.Errorf(".git file has wrong content: %q", string(content))
	}

	// Verify fetch refspec
	cmd := exec.Command("git", "-C", targetDir, "config", "remote.origin.fetch")
	output, err := cmd.Output()
	if err != nil {
		t.Errorf("failed to get fetch refspec: %v", err)
	}
	refspec := strings.TrimSpace(string(output))
	if refspec != "+refs/heads/*:refs/remotes/origin/*" {
		t.Errorf("wrong fetch refspec: %q", refspec)
	}

	// Verify reflog enabled
	cmd = exec.Command("git", "-C", targetDir, "config", "core.logallrefupdates")
	output, err = cmd.Output()
	if err != nil {
		t.Errorf("failed to get reflog config: %v", err)
	}
	reflog := strings.TrimSpace(string(output))
	if reflog != "true" {
		t.Errorf("reflog not enabled: %q", reflog)
	}

	// Verify initial worktree exists
	mainDir := filepath.Join(targetDir, "main")
	if _, err := os.Stat(mainDir); err != nil {
		t.Errorf("initial worktree not created: %v", err)
	}

	// Verify worktree is valid
	cmd = exec.Command("git", "-C", mainDir, "status")
	if err := cmd.Run(); err != nil {
		t.Errorf("git status failed in worktree: %v", err)
	}
}

func TestMkBareRepo_CopiesPttconfig(t *testing.T) {
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	repoDir := createNormalRepoWithConfig(t)
	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Execute mk-bare-repo
	if err := mkBareRepoCmd.RunE(mkBareRepoCmd, []string{}); err != nil {
		t.Fatalf("mk-bare-repo failed: %v", err)
	}

	// Verify .pttconfig copied
	targetDir := filepath.Join(filepath.Dir(repoDir), "repo-bare")
	pttConfigFile := filepath.Join(targetDir, ".pttconfig", "default")
	content, err := os.ReadFile(pttConfigFile)
	if err != nil {
		t.Errorf("failed to read .pttconfig/default: %v", err)
	}
	if string(content) != "copy .env\n" {
		t.Errorf("wrong .pttconfig content: %q", string(content))
	}
}

func TestMkBareRepo_ErrorAlreadyBareRepo(t *testing.T) {
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	bareRepo := createPttBareRepoManual(t)
	mainDir := filepath.Join(bareRepo, "main")
	if err := os.Chdir(mainDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Execute mk-bare-repo
	err := mkBareRepoCmd.RunE(mkBareRepoCmd, []string{})
	if err == nil {
		t.Fatalf("expected error but got none")
	}

	if !strings.Contains(err.Error(), "already in a ptt bare repo") {
		t.Errorf("wrong error message: %v", err)
	}
}

func TestMkBareRepo_ErrorNoRemote(t *testing.T) {
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "repo")
	if err := os.Mkdir(repoDir, 0755); err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	if err := execTestCmd("git", "-C", repoDir, "init", "-b", "main"); err != nil {
		t.Skip("git not available")
	}

	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Execute mk-bare-repo
	err := mkBareRepoCmd.RunE(mkBareRepoCmd, []string{})
	if err == nil {
		t.Fatalf("expected error but got none")
	}

	if !strings.Contains(err.Error(), "no remote origin") {
		t.Errorf("wrong error message: %v", err)
	}
}

func TestMkBareRepo_ErrorTargetExists(t *testing.T) {
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	repoDir := createNormalRepo(t)

	// Create target directory manually
	targetDir := filepath.Join(filepath.Dir(repoDir), "repo-bare")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Execute mk-bare-repo
	err := mkBareRepoCmd.RunE(mkBareRepoCmd, []string{})
	if err == nil {
		t.Fatalf("expected error but got none")
	}

	if !strings.Contains(err.Error(), "target directory already exists") {
		t.Errorf("wrong error message: %v", err)
	}
}

func TestMkBareRepo_FetchWorks(t *testing.T) {
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "repo")
	remoteDir := filepath.Join(tmpDir, "remote.git")

	// Create repo with two branches
	if err := os.Mkdir(repoDir, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	if err := execTestCmd("git", "-C", repoDir, "init", "-b", "main"); err != nil {
		t.Skip("git not available")
	}

	execTestCmd("git", "-C", repoDir, "config", "user.email", "test@example.com")
	execTestCmd("git", "-C", repoDir, "config", "user.name", "Test User")
	execTestCmd("git", "-C", repoDir, "commit", "--allow-empty", "-m", "init")

	// Create second branch
	execTestCmd("git", "-C", repoDir, "checkout", "-b", "feature")
	execTestCmd("git", "-C", repoDir, "commit", "--allow-empty", "-m", "feature commit")
	execTestCmd("git", "-C", repoDir, "checkout", "main")

	// Create remote and push
	execTestCmd("git", "init", "--bare", remoteDir)
	execTestCmd("git", "-C", repoDir, "remote", "add", "origin", remoteDir)
	execTestCmd("git", "-C", repoDir, "push", "-u", "origin", "main")
	execTestCmd("git", "-C", repoDir, "push", "-u", "origin", "feature")

	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Execute mk-bare-repo
	if err := mkBareRepoCmd.RunE(mkBareRepoCmd, []string{}); err != nil {
		t.Fatalf("mk-bare-repo failed: %v", err)
	}

	targetDir := filepath.Join(tmpDir, "repo-bare")

	// Verify fetch works
	cmd := exec.Command("git", "-C", targetDir, "fetch", "origin")
	if err := cmd.Run(); err != nil {
		t.Errorf("git fetch failed: %v", err)
	}

	// Verify both branches are present
	bareDir := filepath.Join(targetDir, ".bare")
	cmd = exec.Command("git", "-C", bareDir, "show-ref")
	output, err := cmd.Output()
	if err != nil {
		t.Errorf("git show-ref failed: %v", err)
	}

	refs := string(output)
	if !strings.Contains(refs, "refs/heads/main") {
		t.Errorf("main branch not found in refs")
	}
	if !strings.Contains(refs, "refs/heads/feature") {
		t.Errorf("feature branch not found in refs")
	}
}
