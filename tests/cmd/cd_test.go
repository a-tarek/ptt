package cmd_test

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCdToHome(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := realPath(t, filepath.Join(containerRoot, "main"))
	stagingPath := addWorktree(t, containerRoot, "staging")

	res := runPtt(t, stagingPath, "cd", "--output-path")
	if res.Err != nil {
		t.Fatalf("cd home failed: %s", res.Stderr)
	}

	got := strings.TrimSpace(res.Stdout)
	if got != mainPath {
		t.Errorf("expected home path %s, got %s", mainPath, got)
	}
}

func TestCdToNamedWorktree(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")
	stagingPath := realPath(t, addWorktree(t, containerRoot, "staging"))

	res := runPtt(t, mainPath, "cd", "--output-path", "staging")
	if res.Err != nil {
		t.Fatalf("cd to staging failed: %s", res.Stderr)
	}

	got := strings.TrimSpace(res.Stdout)
	if got != stagingPath {
		t.Errorf("expected %s, got %s", stagingPath, got)
	}
}

func TestCdFromBareRoot(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	addWorktree(t, containerRoot, "staging")

	res := runPtt(t, containerRoot, "cd", "--output-path", "staging")
	if res.Err != nil {
		t.Fatalf("cd from bare root failed: %s", res.Stderr)
	}

	got := strings.TrimSpace(res.Stdout)
	if got == "" {
		t.Errorf("expected path output, got empty")
	}
}

func TestCdNonexistentWorktree(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")

	res := runPtt(t, mainPath, "cd", "nonexistent")
	if res.Err == nil {
		t.Errorf("expected error for nonexistent worktree")
	}
}

func TestCdAlreadyHome(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")

	res := runPtt(t, mainPath, "cd")
	if res.Err != nil {
		t.Fatalf("cd home from home should succeed: %s", res.Stderr)
	}
}

func TestCdToSelf(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	stagingPath := addWorktree(t, containerRoot, "staging")

	res := runPtt(t, stagingPath, "cd", "--output-path", "staging")
	// Should succeed (already in target is non-error)
	if res.Err != nil {
		t.Fatalf("cd to self should succeed: %s", res.Stderr)
	}
}
