package cmd_test

import (
	"os"
	"path/filepath"
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
