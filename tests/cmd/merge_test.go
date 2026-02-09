package cmd_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeWorktreeBranch(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")
	featurePath := addWorktree(t, containerRoot, "feature")

	// Add a commit in the feature worktree
	os.WriteFile(filepath.Join(featurePath, "feature.txt"), []byte("feature content\n"), 0644)
	gitCmd(t, featurePath, "add", "feature.txt")
	gitCmd(t, featurePath, "commit", "-m", "add feature file")

	// Merge feature into main
	res := runPtt(t, mainPath, "merge", "feature")
	if res.Err != nil {
		t.Fatalf("merge failed: %s", res.Stderr)
	}

	// Verify main now has the feature file
	if _, err := os.Stat(filepath.Join(mainPath, "feature.txt")); os.IsNotExist(err) {
		t.Errorf("feature.txt should exist in main after merge")
	}

	// Verify file content
	content, err := os.ReadFile(filepath.Join(mainPath, "feature.txt"))
	if err != nil {
		t.Fatalf("failed to read feature.txt: %v", err)
	}
	if string(content) != "feature content\n" {
		t.Errorf("unexpected content: %s", content)
	}
}

func TestMergeMultipleCommits(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")
	featurePath := addWorktree(t, containerRoot, "feature")

	// Add multiple commits in feature
	os.WriteFile(filepath.Join(featurePath, "a.txt"), []byte("a\n"), 0644)
	gitCmd(t, featurePath, "add", "a.txt")
	gitCmd(t, featurePath, "commit", "-m", "add a")

	os.WriteFile(filepath.Join(featurePath, "b.txt"), []byte("b\n"), 0644)
	gitCmd(t, featurePath, "add", "b.txt")
	gitCmd(t, featurePath, "commit", "-m", "add b")

	res := runPtt(t, mainPath, "merge", "feature")
	if res.Err != nil {
		t.Fatalf("merge failed: %s", res.Stderr)
	}

	// Both files should exist in main
	if _, err := os.Stat(filepath.Join(mainPath, "a.txt")); os.IsNotExist(err) {
		t.Errorf("a.txt should exist in main after merge")
	}
	if _, err := os.Stat(filepath.Join(mainPath, "b.txt")); os.IsNotExist(err) {
		t.Errorf("b.txt should exist in main after merge")
	}
}

func TestMergeNonexistentWorktree(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")

	res := runPtt(t, mainPath, "merge", "nonexistent")
	if res.Err == nil {
		t.Errorf("expected error for nonexistent worktree")
	}
}

func TestMergeFromBareRoot(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")
	featurePath := addWorktree(t, containerRoot, "feature")

	// Add commit in feature
	os.WriteFile(filepath.Join(featurePath, "feature.txt"), []byte("feature\n"), 0644)
	gitCmd(t, featurePath, "add", "feature.txt")
	gitCmd(t, featurePath, "commit", "-m", "add feature file")

	// Merge from main worktree (since merge needs to run git merge in a worktree)
	res := runPtt(t, mainPath, "merge", "feature")
	if res.Err != nil {
		t.Fatalf("merge from main failed: %s", res.Stderr)
	}

	if _, err := os.Stat(filepath.Join(mainPath, "feature.txt")); os.IsNotExist(err) {
		t.Errorf("feature.txt should exist after merge")
	}
}
