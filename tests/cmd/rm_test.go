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

func TestRmWithYAMLRemoveHook(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")
	stagingPath := addWorktree(t, containerRoot, "staging")

	// Create YAML config with remove hook that writes a marker file
	markerPath := filepath.Join(containerRoot, "hook-marker")
	pttconfigDir := filepath.Join(containerRoot, ".pttconfig")
	os.MkdirAll(pttconfigDir, 0755)
	os.WriteFile(filepath.Join(pttconfigDir, "default.yml"),
		[]byte("remove:\n  - run: touch "+markerPath+"\n"), 0644)

	res := runPtt(t, mainPath, "rm", "staging")
	if res.Err != nil {
		t.Fatalf("rm with YAML remove hook failed: %s", res.Stderr)
	}

	// Marker file should exist (hook ran)
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Error("marker file should exist after remove hook")
	}

	// Worktree should be gone
	if _, err := os.Stat(stagingPath); !os.IsNotExist(err) {
		t.Error("worktree should be removed")
	}
}

func TestRmHookFailureBlocksDeletion(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")
	stagingPath := addWorktree(t, containerRoot, "staging")

	// Create YAML config with a failing remove hook
	pttconfigDir := filepath.Join(containerRoot, ".pttconfig")
	os.MkdirAll(pttconfigDir, 0755)
	os.WriteFile(filepath.Join(pttconfigDir, "default.yml"),
		[]byte("remove:\n  - run: exit 1\n"), 0644)

	res := runPtt(t, mainPath, "rm", "staging")
	if res.Err == nil {
		t.Fatal("expected error when hook fails, got nil")
	}

	// Worktree should still exist
	if _, err := os.Stat(stagingPath); os.IsNotExist(err) {
		t.Error("worktree should still exist after hook failure")
	}
}

func TestRmHookFailureWithForce(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")
	stagingPath := addWorktree(t, containerRoot, "staging")

	// Create YAML config with a failing remove hook
	pttconfigDir := filepath.Join(containerRoot, ".pttconfig")
	os.MkdirAll(pttconfigDir, 0755)
	os.WriteFile(filepath.Join(pttconfigDir, "default.yml"),
		[]byte("remove:\n  - run: exit 1\n"), 0644)

	res := runPtt(t, mainPath, "rm", "--force", "staging")
	if res.Err != nil {
		t.Fatalf("rm --force should succeed despite hook failure: %s", res.Stderr)
	}

	// Worktree should be gone
	if _, err := os.Stat(stagingPath); !os.IsNotExist(err) {
		t.Error("worktree should be removed with --force")
	}
}

func TestRmWithInlineRunFlag(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")
	addWorktree(t, containerRoot, "staging")

	markerPath := filepath.Join(containerRoot, "inline-marker")

	res := runPtt(t, mainPath, "rm", "--run", "touch "+markerPath, "staging")
	if res.Err != nil {
		t.Fatalf("rm --run failed: %s", res.Stderr)
	}

	// Marker file should exist
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Error("marker file should exist after inline --run hook")
	}

	// Worktree should be gone
	stagingPath := filepath.Join(containerRoot, "staging")
	if _, err := os.Stat(stagingPath); !os.IsNotExist(err) {
		t.Error("worktree should be removed")
	}
}

func TestRmWithTextConfigNoHooks(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")
	stagingPath := addWorktree(t, containerRoot, "staging")

	// Create a text config (no remove section possible)
	pttconfigDir := filepath.Join(containerRoot, ".pttconfig")
	os.MkdirAll(pttconfigDir, 0755)
	os.WriteFile(filepath.Join(pttconfigDir, "default"), []byte("copy .env\n"), 0644)

	res := runPtt(t, mainPath, "rm", "staging")
	if res.Err != nil {
		t.Fatalf("rm with text config should succeed: %s", res.Stderr)
	}

	// Worktree should be gone
	if _, err := os.Stat(stagingPath); !os.IsNotExist(err) {
		t.Error("worktree should be removed")
	}
}

func TestRmSkipConfig(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")
	addWorktree(t, containerRoot, "staging")

	// Create YAML config with a failing remove hook
	pttconfigDir := filepath.Join(containerRoot, ".pttconfig")
	os.MkdirAll(pttconfigDir, 0755)
	os.WriteFile(filepath.Join(pttconfigDir, "default.yml"),
		[]byte("remove:\n  - run: exit 1\n"), 0644)

	// --skip-config should skip the failing hook
	res := runPtt(t, mainPath, "rm", "--skip-config", "staging")
	if res.Err != nil {
		t.Fatalf("rm --skip-config should succeed: %s", res.Stderr)
	}

	// Worktree should be gone
	stagingPath := filepath.Join(containerRoot, "staging")
	if _, err := os.Stat(stagingPath); !os.IsNotExist(err) {
		t.Error("worktree should be removed")
	}
}

func TestRmWithNamedConfig(t *testing.T) {
	containerRoot := setupPttBareRepoWithCommit(t)
	mainPath := filepath.Join(containerRoot, "main")
	addWorktree(t, containerRoot, "staging")

	// Create named YAML config with remove hook
	markerPath := filepath.Join(containerRoot, "ci-marker")
	pttconfigDir := filepath.Join(containerRoot, ".pttconfig")
	os.MkdirAll(pttconfigDir, 0755)
	os.WriteFile(filepath.Join(pttconfigDir, "ci.yml"),
		[]byte("remove:\n  - run: touch "+markerPath+"\n"), 0644)

	res := runPtt(t, mainPath, "rm", "--config", "ci", "staging")
	if res.Err != nil {
		t.Fatalf("rm --config ci failed: %s", res.Stderr)
	}

	// Marker from named config should exist
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Error("marker file should exist from named config hook")
	}

	// Worktree should be gone
	stagingPath := filepath.Join(containerRoot, "staging")
	if _, err := os.Stat(stagingPath); !os.IsNotExist(err) {
		t.Error("worktree should be removed")
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
