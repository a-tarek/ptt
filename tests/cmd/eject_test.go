package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEjectFromHomeWorktree(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")

	// Create and switch to feature branch
	gitCmd(t, mainPath, "checkout", "-b", "feature-x")

	// Create a dirty file
	os.WriteFile(filepath.Join(mainPath, "dirty.txt"), []byte("uncommitted\n"), 0644)

	res := runPtt(t, mainPath, "eject")
	if res.Err != nil {
		t.Fatalf("eject failed: %s", res.Stderr)
	}

	// Verify new worktree exists
	expectedPath := filepath.Join(containerRoot, "feature-x")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("new worktree not created at %s", expectedPath)
	}

	// Verify main is back on "main" branch
	branch := gitBranch(t, mainPath)
	if branch != "main" {
		t.Errorf("main should be on 'main', got '%s'", branch)
	}

	// Verify dirty file was carried to new worktree
	if _, err := os.Stat(filepath.Join(expectedPath, "dirty.txt")); os.IsNotExist(err) {
		t.Errorf("dirty file not found in new worktree (stash was not popped)")
	}
}

func TestEjectWithCustomName(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")

	gitCmd(t, mainPath, "checkout", "-b", "feature-x")

	res := runPtt(t, mainPath, "eject", "my-feature")
	if res.Err != nil {
		t.Fatalf("eject with custom name failed: %s", res.Stderr)
	}

	// Should use custom name, not branch name
	expectedPath := filepath.Join(containerRoot, "my-feature")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("new worktree not created at %s (should use custom name)", expectedPath)
	}

	wrongPath := filepath.Join(containerRoot, "feature-x")
	if _, err := os.Stat(wrongPath); err == nil {
		t.Errorf("worktree created at %s (should be my-feature)", wrongPath)
	}
}

func TestEjectDetachedHeadError(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")

	gitCmd(t, mainPath, "checkout", "--detach")

	res := runPtt(t, mainPath, "eject")
	if res.Err == nil {
		t.Fatal("eject should fail on detached HEAD")
	}

	if !strings.Contains(res.Stderr, "detached HEAD") {
		t.Errorf("error should mention 'detached HEAD', got: %s", res.Stderr)
	}
}

func TestEjectAlreadyOnFallbackError(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")

	// Already on "main" (the fallback branch)
	res := runPtt(t, mainPath, "eject")
	if res.Err == nil {
		t.Fatal("eject should fail when already on fallback branch")
	}

	if !strings.Contains(res.Stderr, "already on 'main'") {
		t.Errorf("error should mention already on 'main', got: %s", res.Stderr)
	}
}
