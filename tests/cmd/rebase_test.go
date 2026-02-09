package cmd_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRebaseOntoWorktree(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")
	featurePath := addWorktree(t, containerRoot, "feature")

	// Add commit in main
	os.WriteFile(filepath.Join(mainPath, "main-update.txt"), []byte("update\n"), 0644)
	gitCmd(t, mainPath, "add", "main-update.txt")
	gitCmd(t, mainPath, "commit", "-m", "main update")

	// Rebase feature onto main
	res := runPtt(t, featurePath, "rebase", "main")
	if res.Err != nil {
		t.Fatalf("rebase failed: %s", res.Stderr)
	}

	// Verify feature now has the file from main
	if _, err := os.Stat(filepath.Join(featurePath, "main-update.txt")); os.IsNotExist(err) {
		t.Errorf("main-update.txt should exist in feature after rebase")
	}
}

func TestRebasePreservesFeatureCommits(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")
	featurePath := addWorktree(t, containerRoot, "feature")

	// Add commit in main
	os.WriteFile(filepath.Join(mainPath, "main-update.txt"), []byte("update\n"), 0644)
	gitCmd(t, mainPath, "add", "main-update.txt")
	gitCmd(t, mainPath, "commit", "-m", "main update")

	// Add commit in feature
	os.WriteFile(filepath.Join(featurePath, "feature.txt"), []byte("feature\n"), 0644)
	gitCmd(t, featurePath, "add", "feature.txt")
	gitCmd(t, featurePath, "commit", "-m", "feature commit")

	// Rebase feature onto main
	res := runPtt(t, featurePath, "rebase", "main")
	if res.Err != nil {
		t.Fatalf("rebase failed: %s", res.Stderr)
	}

	// Verify feature has both files
	if _, err := os.Stat(filepath.Join(featurePath, "main-update.txt")); os.IsNotExist(err) {
		t.Errorf("main-update.txt should exist in feature after rebase")
	}
	if _, err := os.Stat(filepath.Join(featurePath, "feature.txt")); os.IsNotExist(err) {
		t.Errorf("feature.txt should still exist in feature after rebase")
	}
}

func TestRebaseNonexistentWorktree(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	featurePath := addWorktree(t, containerRoot, "feature")

	res := runPtt(t, featurePath, "rebase", "nonexistent")
	if res.Err == nil {
		t.Errorf("expected error for nonexistent worktree")
	}
}
