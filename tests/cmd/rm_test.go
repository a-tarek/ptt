package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
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

	if !strings.Contains(res.Stderr, "can't delete current worktree") {
		t.Errorf("expected 'can't delete current worktree' error, got: %s", res.Stderr)
	}

	if _, err := os.Stat(stagingPath); os.IsNotExist(err) {
		t.Errorf("worktree directory should NOT be removed")
	}
}
