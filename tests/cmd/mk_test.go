package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMkCreatesWorktree(t *testing.T) {
	bareRoot := setupStdBareRepo(t)
	mainPath := filepath.Join(bareRoot, "main")

	res := runPtt(t, mainPath, "mk", "staging")
	if res.Err != nil {
		t.Fatalf("mk failed: %s", res.Stderr)
	}

	stagingPath := filepath.Join(bareRoot, "staging")
	if _, err := os.Stat(stagingPath); os.IsNotExist(err) {
		t.Errorf("worktree directory not created at %s", stagingPath)
	}

	if !worktreeListContains(t, mainPath, "staging") {
		t.Errorf("branch 'staging' was not created")
	}
}

func TestMkExistingBranch(t *testing.T) {
	bareRoot := setupStdBareRepo(t)
	mainPath := filepath.Join(bareRoot, "main")

	// Create branch first
	gitCmd(t, mainPath, "branch", "feature")

	res := runPtt(t, mainPath, "mk", "feature")
	if res.Err != nil {
		t.Fatalf("mk with existing branch failed: %s", res.Stderr)
	}

	featurePath := filepath.Join(bareRoot, "feature")
	if _, err := os.Stat(featurePath); os.IsNotExist(err) {
		t.Errorf("worktree directory not created at %s", featurePath)
	}
}

func TestMkWithSkipConfig(t *testing.T) {
	bareRoot := setupStdBareRepo(t)
	mainPath := filepath.Join(bareRoot, "main")

	// Create .wtconfig with copy action
	os.WriteFile(filepath.Join(bareRoot, ".wtconfig"), []byte("copy .env\n"), 0644)

	// Create .env in main worktree
	os.WriteFile(filepath.Join(mainPath, ".env"), []byte("TEST=1\n"), 0644)

	res := runPtt(t, mainPath, "mk", "--skip-config", "staging")
	if res.Err != nil {
		t.Fatalf("mk --skip-config failed: %s", res.Stderr)
	}

	// .env should NOT be copied
	stagingEnv := filepath.Join(bareRoot, "staging", ".env")
	if _, err := os.Stat(stagingEnv); err == nil {
		t.Errorf(".env should not have been copied with --skip-config")
	}
}

func TestMkAlreadyExists(t *testing.T) {
	bareRoot := setupStdBareRepo(t)
	mainPath := filepath.Join(bareRoot, "main")

	// Pre-create target directory
	os.MkdirAll(filepath.Join(bareRoot, "staging"), 0755)

	res := runPtt(t, mainPath, "mk", "staging")
	if res.Err == nil {
		t.Errorf("expected error when path already exists, got nil")
	}
}

func TestMkFromBareRoot(t *testing.T) {
	containerRoot := setupPttBareRepo(t)

	res := runPtt(t, containerRoot, "mk", "staging")
	if res.Err != nil {
		t.Fatalf("mk from bare root should succeed, got: %s", res.Stderr)
	}

	stagingPath := filepath.Join(containerRoot, "staging")
	if _, err := os.Stat(stagingPath); os.IsNotExist(err) {
		t.Errorf("worktree directory not created at %s", stagingPath)
	}

	if !worktreeListContains(t, stagingPath, "staging") {
		t.Errorf("branch 'staging' was not created")
	}
}

func TestMkAppliesCopyConfig(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")

	// Create .pttconfig/default with copy action
	pttconfigDir := filepath.Join(containerRoot, ".pttconfig")
	os.MkdirAll(pttconfigDir, 0755)
	os.WriteFile(filepath.Join(pttconfigDir, "default"), []byte("copy .env\n"), 0644)

	// Create source file in main worktree
	os.WriteFile(filepath.Join(mainPath, ".env"), []byte("DB_HOST=localhost\n"), 0644)

	res := runPtt(t, mainPath, "mk", "staging")
	if res.Err != nil {
		t.Fatalf("mk with config failed: %s", res.Stderr)
	}

	// Verify .env was copied to new worktree
	stagingEnv := filepath.Join(containerRoot, "staging", ".env")
	content, err := os.ReadFile(stagingEnv)
	if err != nil {
		t.Fatalf(".env not copied to staging: %v", err)
	}
	if string(content) != "DB_HOST=localhost\n" {
		t.Errorf("copied .env has wrong content: %s", content)
	}
}

func TestMkWithCopyFlag(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")

	// Create source file
	os.WriteFile(filepath.Join(mainPath, ".env"), []byte("SECRET=abc\n"), 0644)

	res := runPtt(t, mainPath, "mk", "--copy", ".env", "staging")
	if res.Err != nil {
		t.Fatalf("mk --copy failed: %s", res.Stderr)
	}

	// Verify .env was copied
	stagingEnv := filepath.Join(containerRoot, "staging", ".env")
	content, err := os.ReadFile(stagingEnv)
	if err != nil {
		t.Fatalf(".env not copied: %v", err)
	}
	if string(content) != "SECRET=abc\n" {
		t.Errorf("copied file has wrong content: %s", content)
	}
}

func TestMkWithSymlinkFlag(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")

	// Create source directory
	os.MkdirAll(filepath.Join(mainPath, "node_modules", "pkg"), 0755)
	os.WriteFile(filepath.Join(mainPath, "node_modules", "pkg", "index.js"), []byte("//\n"), 0644)

	res := runPtt(t, mainPath, "mk", "--symlink", "node_modules", "staging")
	if res.Err != nil {
		t.Fatalf("mk --symlink failed: %s", res.Stderr)
	}

	// Verify node_modules is a symlink in staging
	stagingNM := filepath.Join(containerRoot, "staging", "node_modules")
	fi, err := os.Lstat(stagingNM)
	if err != nil {
		t.Fatalf("node_modules not created in staging: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("node_modules should be a symlink")
	}
}

func TestMkWorktreeOnCorrectBranch(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")

	res := runPtt(t, mainPath, "mk", "feature-abc")
	if res.Err != nil {
		t.Fatalf("mk failed: %s", res.Stderr)
	}

	featurePath := filepath.Join(containerRoot, "feature-abc")
	branch := gitBranch(t, featurePath)
	if branch != "feature-abc" {
		t.Errorf("expected branch 'feature-abc', got '%s'", branch)
	}
}

func TestMkOutputPath(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")

	res := runPtt(t, mainPath, "mk", "--output-path", "staging")
	if res.Err != nil {
		t.Fatalf("mk --output-path failed: %s", res.Stderr)
	}

	expectedPath := realPath(t, filepath.Join(containerRoot, "staging"))
	got := strings.TrimSpace(res.Stdout)
	if got != expectedPath {
		t.Errorf("expected output path %s, got %s", expectedPath, got)
	}
}
