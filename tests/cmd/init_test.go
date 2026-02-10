package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit_NormalRepo_Success(t *testing.T) {
	repoDir := setupNormalRepoWithRemote(t)

	res := runPtt(t, repoDir, "init", "-y")
	if res.Err != nil {
		t.Fatalf("init failed: %s", res.Stderr)
	}

	// Verify .bare/ created
	if _, err := os.Stat(filepath.Join(repoDir, ".bare")); os.IsNotExist(err) {
		t.Errorf(".bare/ directory not created")
	}

	// Verify .git pointer file
	content, err := os.ReadFile(filepath.Join(repoDir, ".git"))
	if err != nil {
		t.Fatalf("failed to read .git pointer: %v", err)
	}
	if strings.TrimSpace(string(content)) != "gitdir: ./.bare" {
		t.Errorf("incorrect .git pointer: %s", content)
	}

	// Verify default branch worktree exists
	headRef := gitCmd(t, repoDir, "symbolic-ref", "HEAD")
	defaultBranch := strings.TrimPrefix(headRef, "refs/heads/")
	if _, err := os.Stat(filepath.Join(repoDir, defaultBranch)); os.IsNotExist(err) {
		t.Errorf("%s worktree not created", defaultBranch)
	}

	// Verify .pttconfig/default created
	if _, err := os.Stat(filepath.Join(repoDir, ".pttconfig", "default.yml")); os.IsNotExist(err) {
		t.Errorf(".pttconfig/default.yml not created")
	}
}

func TestInit_NormalRepo_FeatureBranch(t *testing.T) {
	repoDir := setupNormalRepoWithRemote(t)
	gitCmd(t, repoDir, "checkout", "-b", "feature/test")

	res := runPtt(t, repoDir, "init", "-y")
	if res.Err != nil {
		t.Fatalf("init failed: %s", res.Stderr)
	}

	porcelain := gitCmd(t, repoDir, "worktree", "list", "--porcelain")

	if !strings.Contains(porcelain, "branch refs/heads/master") {
		t.Errorf("master worktree not created")
	}
	if !strings.Contains(porcelain, "branch refs/heads/feature/test") {
		t.Errorf("feature/test worktree not created")
	}
}

func TestInit_NormalRepo_PreservesUntracked(t *testing.T) {
	repoDir := setupNormalRepoWithRemote(t)

	// Create untracked file
	os.WriteFile(filepath.Join(repoDir, "untracked.txt"), []byte("test"), 0644)

	res := runPtt(t, repoDir, "init", "-y")
	if res.Err != nil {
		t.Fatalf("init failed: %s", res.Stderr)
	}

	// Find the default branch worktree
	headRef := gitCmd(t, repoDir, "symbolic-ref", "HEAD")
	defaultBranch := strings.TrimPrefix(headRef, "refs/heads/")
	untrackedInWT := filepath.Join(repoDir, defaultBranch, "untracked.txt")

	if _, err := os.Stat(untrackedInWT); os.IsNotExist(err) {
		t.Errorf("untracked file was not preserved in worktree")
	}
}

func TestInit_NormalRepo_DirtyRejects(t *testing.T) {
	repoDir := setupNormalRepoWithRemote(t)

	// Make staged change (dirty)
	os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("# Modified\n"), 0644)

	res := runPtt(t, repoDir, "init", "-y")
	if res.Err == nil {
		t.Errorf("expected error for dirty working tree")
	}
}

func TestInit_NormalRepo_LocalOnly(t *testing.T) {
	repoDir := setupLocalOnlyRepo(t)

	res := runPtt(t, repoDir, "init", "-y")
	if res.Err != nil {
		t.Fatalf("init failed: %s", res.Stderr)
	}

	if _, err := os.Stat(filepath.Join(repoDir, ".pttconfig", "default.yml")); os.IsNotExist(err) {
		t.Errorf(".pttconfig/default.yml not created")
	}
}

func TestInit_RawBareRepo_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source repo
	sourceDir := filepath.Join(tmpDir, "source")
	os.MkdirAll(sourceDir, 0755)
	gitCmd(t, sourceDir, "init")
	gitCmd(t, sourceDir, "config", "user.email", "test@example.com")
	gitCmd(t, sourceDir, "config", "user.name", "Test User")
	os.WriteFile(filepath.Join(sourceDir, "README.md"), []byte("# Test\n"), 0644)
	gitCmd(t, sourceDir, "add", "README.md")
	gitCmd(t, sourceDir, "commit", "-m", "Initial commit")

	// Clone as bare
	bareDir := filepath.Join(tmpDir, "repo.git")
	gitCmd(t, tmpDir, "clone", "--bare", sourceDir, bareDir)

	res := runPtt(t, bareDir, "init", "-y")
	if res.Err != nil {
		t.Fatalf("init failed: %s", res.Stderr)
	}

	// Verify .bare/ wrapper
	if _, err := os.Stat(filepath.Join(bareDir, ".bare")); os.IsNotExist(err) {
		t.Errorf(".bare/ directory not created")
	}

	// Verify .git pointer
	content, err := os.ReadFile(filepath.Join(bareDir, ".git"))
	if err != nil {
		t.Fatalf("failed to read .git: %v", err)
	}
	if strings.TrimSpace(string(content)) != "gitdir: ./.bare" {
		t.Errorf("incorrect .git pointer: %s", content)
	}

	// Verify master worktree
	if _, err := os.Stat(filepath.Join(bareDir, "master")); os.IsNotExist(err) {
		t.Errorf("master worktree not created")
	}

	// Verify reflog enabled
	logAllRef := gitCmd(t, bareDir, "config", "core.logallrefupdates")
	if logAllRef != "true" {
		t.Errorf("reflog not enabled")
	}
}

func TestInit_PttBare_CreatesPttconfig(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)

	res := runPtt(t, containerRoot, "init", "-y")
	if res.Err != nil {
		t.Fatalf("init failed: %s", res.Stderr)
	}

	if _, err := os.Stat(filepath.Join(containerRoot, ".pttconfig", "default.yml")); os.IsNotExist(err) {
		t.Errorf(".pttconfig/default.yml not created")
	}
}

func TestInit_PttBare_AlreadySetUp(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)

	// Pre-create .pttconfig/default.yml
	pttconfigDir := filepath.Join(containerRoot, ".pttconfig")
	os.MkdirAll(pttconfigDir, 0755)
	os.WriteFile(filepath.Join(pttconfigDir, "default.yml"), []byte(""), 0644)

	res := runPtt(t, containerRoot, "init", "-y")
	if res.Err != nil {
		t.Errorf("init should be idempotent, got: %s", res.Stderr)
	}
}

func TestInit_CreatesDefaultYAMLWithStub(t *testing.T) {
	repoDir := setupNormalRepoWithRemote(t)

	res := runPtt(t, repoDir, "init", "-y")
	if res.Err != nil {
		t.Fatalf("init failed: %s", res.Stderr)
	}

	content, err := os.ReadFile(filepath.Join(repoDir, ".pttconfig", "default.yml"))
	if err != nil {
		t.Fatalf("failed to read default.yml: %v", err)
	}

	body := string(content)
	for _, want := range []string{"# create:", "#   - copy:", "#   - symlink:", "#   - run:", "# remove:"} {
		if !strings.Contains(body, want) {
			t.Errorf("default.yml missing expected line %q", want)
		}
	}
}

func TestInit_PttBare_RepairsMissingYAMLConfig(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)

	// Ensure no .pttconfig/ exists
	os.RemoveAll(filepath.Join(containerRoot, ".pttconfig"))

	res := runPtt(t, containerRoot, "init", "-y")
	if res.Err != nil {
		t.Fatalf("init failed: %s", res.Stderr)
	}

	content, err := os.ReadFile(filepath.Join(containerRoot, ".pttconfig", "default.yml"))
	if err != nil {
		t.Fatalf(".pttconfig/default.yml not created: %v", err)
	}

	if !strings.Contains(string(content), "# create:") {
		t.Errorf("default.yml missing stub content")
	}
}

func TestInit_NotGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	res := runPtt(t, tmpDir, "init", "-y")
	if res.Err == nil {
		t.Errorf("expected error for non-git directory")
	}
}
