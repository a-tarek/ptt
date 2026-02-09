package cmd_test

import (
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
