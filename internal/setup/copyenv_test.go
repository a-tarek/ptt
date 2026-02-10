package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/a-tarek/ptt/internal/config"
	"github.com/a-tarek/ptt/internal/git"
)

func TestExecuteCopyEnv_CopiesFile(t *testing.T) {
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	envContent := "PORT=3000\nHOST=localhost\n"
	if err := os.WriteFile(filepath.Join(srcDir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.CopyEnvConfig{
		File: ".env",
		Vars: map[string]config.EnvVarRule{},
	}

	if err := ExecuteCopyEnv(srcDir, targetDir, cfg); err != nil {
		t.Fatalf("ExecuteCopyEnv failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(targetDir, ".env"))
	if err != nil {
		t.Fatalf("target file not created: %v", err)
	}
	if string(got) != envContent {
		t.Errorf("content mismatch: got %q, want %q", got, envContent)
	}
}

func TestExecuteCopyEnv_ShellCommandStrategy(t *testing.T) {
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	envContent := "PORT=3000\nBRANCH=main\n"
	if err := os.WriteFile(filepath.Join(srcDir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.CopyEnvConfig{
		File: ".env",
		Vars: map[string]config.EnvVarRule{
			"BRANCH": {Strategy: "echo feature-1"},
		},
	}

	if err := ExecuteCopyEnv(srcDir, targetDir, cfg); err != nil {
		t.Fatalf("ExecuteCopyEnv failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(targetDir, ".env"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "BRANCH=feature-1") {
		t.Errorf("expected BRANCH=feature-1, got:\n%s", got)
	}
}

func TestExecuteCopyEnv_DirectValueOverride(t *testing.T) {
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	envContent := "PORT=3000\nHOST=localhost\n"
	if err := os.WriteFile(filepath.Join(srcDir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.CopyEnvConfig{
		File:      ".env",
		Vars:      map[string]config.EnvVarRule{},
		Overrides: map[string]string{"PORT": "4000"},
	}

	if err := ExecuteCopyEnv(srcDir, targetDir, cfg); err != nil {
		t.Fatalf("ExecuteCopyEnv failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(targetDir, ".env"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "PORT=4000") {
		t.Errorf("expected PORT=4000, got:\n%s", got)
	}
	// HOST should remain unchanged
	if !strings.Contains(string(got), "HOST=localhost") {
		t.Errorf("expected HOST=localhost unchanged, got:\n%s", got)
	}
}

func TestExecuteCopyEnv_OverrideWinsOverConfig(t *testing.T) {
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	envContent := "PORT=3000\n"
	if err := os.WriteFile(filepath.Join(srcDir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.CopyEnvConfig{
		File: ".env",
		Vars: map[string]config.EnvVarRule{
			"PORT": {Strategy: "echo 9999"},
		},
		Overrides: map[string]string{"PORT": "5000"},
	}

	if err := ExecuteCopyEnv(srcDir, targetDir, cfg); err != nil {
		t.Fatalf("ExecuteCopyEnv failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(targetDir, ".env"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "PORT=5000") {
		t.Errorf("expected override to win, got:\n%s", got)
	}
}

func TestResolveIncrementValue(t *testing.T) {
	// Create fake worktree dirs with env files
	wt1 := t.TempDir()
	wt2 := t.TempDir()
	wt3 := t.TempDir() // missing var

	os.WriteFile(filepath.Join(wt1, ".env"), []byte("PORT=3000\n"), 0644)
	os.WriteFile(filepath.Join(wt2, ".env"), []byte("PORT=3005\n"), 0644)
	os.WriteFile(filepath.Join(wt3, ".env"), []byte("HOST=localhost\n"), 0644)

	worktrees := []git.Worktree{
		{Path: wt1},
		{Path: wt2},
		{Path: wt3},
	}

	result, err := ResolveIncrementValue(worktrees, ".env", "PORT")
	if err != nil {
		t.Fatalf("ResolveIncrementValue failed: %v", err)
	}
	if result != 3006 {
		t.Errorf("expected 3006, got %d", result)
	}
}

func TestResolveIncrementValue_NoWorktreesHaveVar(t *testing.T) {
	wt1 := t.TempDir()
	os.WriteFile(filepath.Join(wt1, ".env"), []byte("HOST=localhost\n"), 0644)

	worktrees := []git.Worktree{{Path: wt1}}

	_, err := ResolveIncrementValue(worktrees, ".env", "PORT")
	if err == nil {
		t.Fatal("expected error when no worktree has the var")
	}
}

func TestRunStrategyCommand(t *testing.T) {
	workDir := t.TempDir()

	result, err := RunStrategyCommand(workDir, "echo hello")
	if err != nil {
		t.Fatalf("RunStrategyCommand failed: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestRunStrategyCommand_Failure(t *testing.T) {
	workDir := t.TempDir()

	_, err := RunStrategyCommand(workDir, "exit 1")
	if err == nil {
		t.Fatal("expected error for failing command")
	}
}
