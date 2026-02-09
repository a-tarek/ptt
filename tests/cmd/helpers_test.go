package cmd_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var (
	pttBinaryPath string
	buildOnce     sync.Once
)

// buildBinary compiles the ptt binary once per test run.
func buildBinary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		tmpDir, err := os.MkdirTemp("", "ptt-cmd-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		pttBinaryPath = filepath.Join(tmpDir, "ptt")
		projectRoot := filepath.Join("..", "..")

		cmd := exec.Command("go", "build", "-o", pttBinaryPath, ".")
		cmd.Dir = projectRoot
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("failed to build ptt: %v\n%s", err, output)
		}
	})
	return pttBinaryPath
}

type pttResult struct {
	Stdout string
	Stderr string
	Err    error
}

// runPtt runs the ptt binary with args in the given directory.
func runPtt(t *testing.T, dir string, args ...string) pttResult {
	t.Helper()
	return runPttWithStdin(t, dir, "", args...)
}

// runPttWithStdin runs the ptt binary with args and stdin input.
func runPttWithStdin(t *testing.T, dir string, stdin string, args ...string) pttResult {
	t.Helper()
	bin := buildBinary(t)
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return pttResult{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
}

// gitCmd runs a git command and fatals on error.
func gitCmd(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output))
}

// setupPttBareRepo creates a ptt-style bare repository with a main worktree.
// Structure: containerRoot/.bare, containerRoot/.git (pointer), containerRoot/main/
func setupPttBareRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Create origin with initial commit
	originDir := filepath.Join(tmpDir, "origin")
	os.Mkdir(originDir, 0755)
	gitCmd(t, originDir, "init")
	gitCmd(t, originDir, "config", "user.email", "test@example.com")
	gitCmd(t, originDir, "config", "user.name", "Test User")
	os.WriteFile(filepath.Join(originDir, "README.md"), []byte("# Test\n"), 0644)
	gitCmd(t, originDir, "add", "README.md")
	gitCmd(t, originDir, "commit", "-m", "Initial commit")

	// Create container with .bare clone
	containerRoot := filepath.Join(tmpDir, "project-bare")
	os.Mkdir(containerRoot, 0755)

	bareDir := filepath.Join(containerRoot, ".bare")
	gitCmd(t, tmpDir, "clone", "--bare", originDir, bareDir)

	// Create .git pointer file
	os.WriteFile(filepath.Join(containerRoot, ".git"), []byte("gitdir: ./.bare\n"), 0644)

	// Configure fetch refspec
	gitCmd(t, containerRoot, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
	gitCmd(t, containerRoot, "fetch", "origin")

	// Create main worktree
	gitCmd(t, bareDir, "worktree", "add", filepath.Join(containerRoot, "main"), "master")

	return containerRoot
}

// setupPttBareRepoWithCommit creates a ptt-style bare repo with .bare/.git pointer
// and a main worktree that has at least one commit. Uses git init --bare (no origin).
func setupPttBareRepoWithCommit(t *testing.T) string {
	t.Helper()

	containerRoot := t.TempDir()
	bareDir := filepath.Join(containerRoot, ".bare")

	gitCmd(t, containerRoot, "init", "--bare", bareDir)
	os.WriteFile(filepath.Join(containerRoot, ".git"), []byte("gitdir: ./.bare\n"), 0644)
	gitCmd(t, containerRoot, "config", "core.bare", "true")
	gitCmd(t, containerRoot, "config", "core.logallrefupdates", "true")

	mainPath := filepath.Join(containerRoot, "main")
	gitCmd(t, containerRoot, "worktree", "add", mainPath, "-b", "main")
	gitCmd(t, mainPath, "config", "user.email", "test@example.com")
	gitCmd(t, mainPath, "config", "user.name", "Test User")
	os.WriteFile(filepath.Join(mainPath, "README.md"), []byte("# Test\n"), 0644)
	gitCmd(t, mainPath, "add", "README.md")
	gitCmd(t, mainPath, "commit", "-m", "Initial commit")

	return containerRoot
}

// setupStdBareRepo creates a standard bare repository (repo.git/) with a main worktree.
func setupStdBareRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	bareRoot := filepath.Join(tmpDir, "repo.git")

	gitCmd(t, tmpDir, "init", "--bare", bareRoot)

	mainPath := filepath.Join(bareRoot, "main")
	gitCmd(t, bareRoot, "worktree", "add", mainPath, "-b", "main")

	os.WriteFile(filepath.Join(mainPath, "README.md"), []byte("# Test\n"), 0644)
	gitCmd(t, mainPath, "config", "user.email", "test@example.com")
	gitCmd(t, mainPath, "config", "user.name", "Test User")
	gitCmd(t, mainPath, "add", "README.md")
	gitCmd(t, mainPath, "commit", "-m", "Initial commit")

	return bareRoot
}

// setupRegularRepo creates a standard (non-bare) git repository.
func setupRegularRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	gitCmd(t, tmpDir, "init")
	gitCmd(t, tmpDir, "config", "user.email", "test@example.com")
	gitCmd(t, tmpDir, "config", "user.name", "Test User")
	os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("test\n"), 0644)
	gitCmd(t, tmpDir, "add", "README.md")
	gitCmd(t, tmpDir, "commit", "-m", "Initial commit")

	return tmpDir
}

// setupNormalRepoWithRemote creates a normal clone with a remote.
func setupNormalRepoWithRemote(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Create bare remote
	remoteDir := filepath.Join(tmpDir, "remote.git")
	gitCmd(t, tmpDir, "init", "--bare", remoteDir)

	// Clone it
	repoDir := filepath.Join(tmpDir, "repo")
	gitCmd(t, tmpDir, "clone", remoteDir, repoDir)
	gitCmd(t, repoDir, "config", "user.email", "test@example.com")
	gitCmd(t, repoDir, "config", "user.name", "Test User")

	os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("# Test\n"), 0644)
	gitCmd(t, repoDir, "add", "README.md")
	gitCmd(t, repoDir, "commit", "-m", "Initial commit")

	// Push to remote
	branch := gitCmd(t, repoDir, "branch", "--show-current")
	gitCmd(t, repoDir, "push", "origin", branch)

	return repoDir
}

// setupLocalOnlyRepo creates a normal repo without a remote.
func setupLocalOnlyRepo(t *testing.T) string {
	t.Helper()

	repoDir := filepath.Join(t.TempDir(), "repo")
	os.MkdirAll(repoDir, 0755)
	gitCmd(t, repoDir, "init")
	gitCmd(t, repoDir, "config", "user.email", "test@example.com")
	gitCmd(t, repoDir, "config", "user.name", "Test User")
	os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("# Test\n"), 0644)
	gitCmd(t, repoDir, "add", "README.md")
	gitCmd(t, repoDir, "commit", "-m", "Initial commit")

	return repoDir
}

// addWorktree creates an additional worktree in a ptt bare repo.
func addWorktree(t *testing.T, containerRoot, name string) string {
	t.Helper()
	bareDir := filepath.Join(containerRoot, ".bare")
	wtPath := filepath.Join(containerRoot, name)
	gitCmd(t, bareDir, "worktree", "add", wtPath, "-b", name)
	return wtPath
}

// gitBranch returns the current branch name in the given directory.
func gitBranch(t *testing.T, dir string) string {
	t.Helper()
	return gitCmd(t, dir, "branch", "--show-current")
}

// worktreeListContains checks if git worktree list output contains a name.
func worktreeListContains(t *testing.T, dir, name string) bool {
	t.Helper()
	output := gitCmd(t, dir, "worktree", "list", "--porcelain")
	return strings.Contains(output, name)
}
