package cmd_test

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLsFiltersBareEntry(t *testing.T) {
	containerRoot := setupPttBareRepo(t)
	mainPath := containerRoot + "/main"

	// Verify git sees the bare entry
	porcelain := gitCmd(t, mainPath, "worktree", "list", "--porcelain")
	if !strings.Contains(porcelain, "bare") {
		t.Skip("git doesn't report bare entry, test environment issue")
	}

	res := runPtt(t, mainPath, "ls")
	if res.Err != nil {
		t.Fatalf("ls failed: %s", res.Stderr)
	}

	// ptt ls should NOT show .bare metadata entry
	if strings.Contains(res.Stdout, ".bare") {
		t.Errorf("output should not contain .bare entry, got:\n%s", res.Stdout)
	}

	if len(strings.TrimSpace(res.Stdout)) == 0 {
		t.Errorf("output should not be empty")
	}

	if !strings.Contains(res.Stdout, "master") {
		t.Errorf("output should contain master branch, got:\n%s", res.Stdout)
	}
}

func TestLsNormalRepo(t *testing.T) {
	repoDir := setupRegularRepo(t)

	res := runPtt(t, repoDir, "ls")
	if res.Err != nil {
		t.Fatalf("ls in normal repo failed: %s", res.Stderr)
	}
}

func TestLsShowsMultipleWorktrees(t *testing.T) {
	containerRoot := setupPttBareRepo(t)
	addWorktree(t, containerRoot, "staging")
	addWorktree(t, containerRoot, "feature")
	mainPath := filepath.Join(containerRoot, "main")

	res := runPtt(t, mainPath, "ls")
	if res.Err != nil {
		t.Fatalf("ls failed: %s", res.Stderr)
	}

	output := res.Stdout
	if !strings.Contains(output, "main") {
		t.Errorf("output should contain 'main' worktree")
	}
	if !strings.Contains(output, "staging") {
		t.Errorf("output should contain 'staging' worktree")
	}
	if !strings.Contains(output, "feature") {
		t.Errorf("output should contain 'feature' worktree")
	}
}

func TestLsFromBareRoot(t *testing.T) {
	containerRoot := setupPttBareRepo(t)

	res := runPtt(t, containerRoot, "ls")
	if res.Err != nil {
		t.Fatalf("ls from bare root failed: %s", res.Stderr)
	}

	if len(strings.TrimSpace(res.Stdout)) == 0 {
		t.Errorf("output should not be empty from bare root")
	}
}

func TestLsShowsAllPaths(t *testing.T) {
	containerRoot := setupPttBareRepo(t)
	mainPath := filepath.Join(containerRoot, "main")

	res := runPtt(t, mainPath, "ls", "--all")
	if res.Err != nil {
		t.Fatalf("ls --all failed: %s", res.Stderr)
	}

	// With --all, output should contain the full path
	if !strings.Contains(res.Stdout, containerRoot) {
		t.Errorf("--all output should contain full paths, got:\n%s", res.Stdout)
	}
}
