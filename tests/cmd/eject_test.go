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
}

func TestEjectAlreadyOnFallbackError(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")

	// Already on "main" (the fallback branch)
	res := runPtt(t, mainPath, "eject")
	if res.Err == nil {
		t.Fatal("eject should fail when already on fallback branch")
	}
}

func TestEjectPreservesTrackedChanges(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")

	// Switch to feature branch
	gitCmd(t, mainPath, "checkout", "-b", "feature-y")

	// Modify a tracked file (creates a dirty state that IsDirty detects)
	os.WriteFile(filepath.Join(mainPath, "README.md"), []byte("# Modified\n"), 0644)

	res := runPtt(t, mainPath, "eject")
	if res.Err != nil {
		t.Fatalf("eject failed: %s", res.Stderr)
	}

	// Verify tracked change was carried to new worktree
	featurePath := filepath.Join(containerRoot, "feature-y")
	content, err := os.ReadFile(filepath.Join(featurePath, "README.md"))
	if err != nil {
		t.Fatalf("README.md not found in ejected worktree: %v", err)
	}
	if string(content) != "# Modified\n" {
		t.Errorf("tracked change not preserved, got: %s", content)
	}

	// Verify main is back on "main" branch
	branch := gitBranch(t, mainPath)
	if branch != "main" {
		t.Errorf("main should be back on 'main', got '%s'", branch)
	}
}

func TestEjectOutputPath(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")

	gitCmd(t, mainPath, "checkout", "-b", "feature-z")

	res := runPtt(t, mainPath, "eject", "--output-path")
	if res.Err != nil {
		t.Fatalf("eject --output-path failed: %s", res.Stderr)
	}

	expectedPath := realPath(t, filepath.Join(containerRoot, "feature-z"))
	got := strings.TrimSpace(res.Stdout)
	if got != expectedPath {
		t.Errorf("expected output path %s, got %s", expectedPath, got)
	}
}
