package cmd_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRmFromOtherWorktree(t *testing.T) {
	containerRoot := setupPttBareRepo(t)
	stagingPath := addWorktree(t, containerRoot, "staging")
	mainPath := filepath.Join(containerRoot, "main")

	res := runPtt(t, mainPath, "rm", "staging")
	if res.Err != nil {
		t.Fatalf("rm from other worktree should succeed, got: %s", res.Stderr)
	}

	if _, err := os.Stat(stagingPath); !os.IsNotExist(err) {
		t.Errorf("worktree directory should be removed: %s", stagingPath)
	}

	if worktreeListContains(t, mainPath, "staging") {
		t.Errorf("git should no longer list staging worktree")
	}
}

func TestRmFromBareRoot(t *testing.T) {
	containerRoot := setupPttBareRepo(t)
	stagingPath := addWorktree(t, containerRoot, "staging")

	res := runPtt(t, containerRoot, "rm", "staging")
	if res.Err != nil {
		t.Fatalf("rm from bare root should succeed, got: %s", res.Stderr)
	}

	if _, err := os.Stat(stagingPath); !os.IsNotExist(err) {
		t.Errorf("worktree directory should be removed: %s", stagingPath)
	}
}

func TestRmSelf(t *testing.T) {
	containerRoot := setupPttBareRepo(t)
	stagingPath := addWorktree(t, containerRoot, "staging")

	res := runPtt(t, stagingPath, "rm", "staging")
	if res.Err == nil {
		t.Fatal("rm self should fail")
	}

	// Worktree should still exist
	if _, err := os.Stat(stagingPath); os.IsNotExist(err) {
		t.Errorf("worktree directory should NOT be removed")
	}
}

func TestRmForce(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")
	stagingPath := addWorktree(t, containerRoot, "staging")

	// Make staging dirty (modify tracked file)
	os.WriteFile(filepath.Join(stagingPath, "README.md"), []byte("# Modified\n"), 0644)

	// rm without --force should fail (no stdin for confirmation)
	res := runPtt(t, mainPath, "rm", "staging")
	if res.Err == nil {
		t.Errorf("expected error when removing dirty worktree without --force")
	}

	// Worktree should still exist
	if _, err := os.Stat(stagingPath); os.IsNotExist(err) {
		t.Errorf("worktree should still exist after failed rm")
	}

	// rm --force should succeed
	res = runPtt(t, mainPath, "rm", "--force", "staging")
	if res.Err != nil {
		t.Fatalf("rm --force failed: %s", res.Stderr)
	}

	if _, err := os.Stat(stagingPath); !os.IsNotExist(err) {
		t.Errorf("worktree directory should be removed after --force")
	}
}

func TestRmWithBranch(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")
	addWorktree(t, containerRoot, "staging")

	res := runPtt(t, mainPath, "rm", "--branch", "staging")
	if res.Err != nil {
		t.Fatalf("rm --branch failed: %s", res.Stderr)
	}

	// Worktree should be gone
	stagingPath := filepath.Join(containerRoot, "staging")
	if _, err := os.Stat(stagingPath); !os.IsNotExist(err) {
		t.Errorf("worktree directory should be removed")
	}

	// Branch should no longer exist
	if branchExists(t, mainPath, "staging") {
		t.Errorf("branch 'staging' should have been deleted")
	}
}

func TestRmNonexistent(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")

	res := runPtt(t, mainPath, "rm", "nonexistent")
	if res.Err == nil {
		t.Errorf("expected error for nonexistent worktree")
	}
}
