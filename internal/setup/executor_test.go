package setup

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ahmedelarabyy/wt/internal/config"
)

func TestExecuteActionsHappyPath(t *testing.T) {
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	// Create source files
	envFile := filepath.Join(srcDir, ".env")
	if err := os.WriteFile(envFile, []byte("ENV=test"), 0644); err != nil {
		t.Fatalf("failed to create .env: %v", err)
	}

	nodeModules := filepath.Join(srcDir, "node_modules")
	if err := os.MkdirAll(nodeModules, 0755); err != nil {
		t.Fatalf("failed to create node_modules: %v", err)
	}

	// Define actions
	actions := []config.Action{
		{Type: config.ActionCopy, Path: ".env"},
		{Type: config.ActionSymlink, Path: "node_modules"},
		{Type: config.ActionRun, Path: "echo done"},
	}

	// Execute
	err := ExecuteActions(srcDir, targetDir, actions)
	if err != nil {
		t.Fatalf("ExecuteActions failed: %v", err)
	}

	// Verify .env was copied
	targetEnv := filepath.Join(targetDir, ".env")
	content, err := os.ReadFile(targetEnv)
	if err != nil {
		t.Errorf(".env not copied: %v", err)
	} else if string(content) != "ENV=test" {
		t.Errorf(".env content mismatch: got %q, want %q", content, "ENV=test")
	}

	// Verify node_modules was symlinked
	targetModules := filepath.Join(targetDir, "node_modules")
	linkDest, err := os.Readlink(targetModules)
	if err != nil {
		t.Errorf("node_modules not symlinked: %v", err)
	} else if linkDest != nodeModules {
		t.Errorf("node_modules symlink incorrect: got %q, want %q", linkDest, nodeModules)
	}
}

func TestExecuteActionsSequentialOrder(t *testing.T) {
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	// Actions that write to a file in sequence
	actions := []config.Action{
		{Type: config.ActionRun, Path: "echo 1 >> order.txt"},
		{Type: config.ActionRun, Path: "echo 2 >> order.txt"},
		{Type: config.ActionRun, Path: "echo 3 >> order.txt"},
	}

	err := ExecuteActions(srcDir, targetDir, actions)
	if err != nil {
		t.Fatalf("ExecuteActions failed: %v", err)
	}

	// Verify order
	orderFile := filepath.Join(targetDir, "order.txt")
	content, err := os.ReadFile(orderFile)
	if err != nil {
		t.Fatalf("order.txt not created: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	want := []string{"1", "2", "3"}
	if len(lines) != len(want) {
		t.Fatalf("wrong number of lines: got %d, want %d", len(lines), len(want))
	}
	for i, line := range lines {
		if line != want[i] {
			t.Errorf("line %d: got %q, want %q", i, line, want[i])
		}
	}
}

func TestExecuteActionsStatusMessages(t *testing.T) {
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	// Create source file
	envFile := filepath.Join(srcDir, ".env")
	if err := os.WriteFile(envFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create .env: %v", err)
	}

	nodeModules := filepath.Join(srcDir, "node_modules")
	if err := os.MkdirAll(nodeModules, 0755); err != nil {
		t.Fatalf("failed to create node_modules: %v", err)
	}

	actions := []config.Action{
		{Type: config.ActionCopy, Path: ".env"},
		{Type: config.ActionSymlink, Path: "node_modules"},
		{Type: config.ActionRun, Path: "echo done"},
	}

	// Capture stderr (status messages go to stderr, not stdout)
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := ExecuteActions(srcDir, targetDir, actions)

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if err != nil {
		t.Fatalf("ExecuteActions failed: %v", err)
	}

	// Verify status messages
	if !strings.Contains(output, "Copied .env") {
		t.Errorf("output missing 'Copied .env': %q", output)
	}
	if !strings.Contains(output, "Symlinked node_modules") {
		t.Errorf("output missing 'Symlinked node_modules': %q", output)
	}
	if !strings.Contains(output, "Running echo done...") {
		t.Errorf("output missing 'Running echo done...': %q", output)
	}
}

func TestExecuteActionsRunFailureTriggersRollback(t *testing.T) {
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	// Create source file to copy
	envFile := filepath.Join(srcDir, ".env")
	if err := os.WriteFile(envFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create .env: %v", err)
	}

	// Actions: copy succeeds, then run fails
	actions := []config.Action{
		{Type: config.ActionCopy, Path: ".env"},
		{Type: config.ActionRun, Path: "exit 1"},
	}

	// Capture stderr to verify rollback message
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Execute - should fail and trigger rollback
	err := ExecuteActions(srcDir, targetDir, actions)

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)
	stderrOutput := buf.String()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Verify rollback was attempted (message printed)
	if !strings.Contains(stderrOutput, "Error occurred, rolling back worktree...") {
		t.Errorf("stderr missing rollback message: %q", stderrOutput)
	}

	// Verify target directory was cleaned up (fallback removal works)
	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		t.Errorf("target directory still exists after rollback")
	}
}

func TestExecuteActionsCopyFailureTriggersRollback(t *testing.T) {
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	// Actions with nonexistent source file
	actions := []config.Action{
		{Type: config.ActionCopy, Path: "nonexistent.txt"},
	}

	err := ExecuteActions(srcDir, targetDir, actions)
	if err == nil {
		t.Error("expected error for missing source, got nil")
	}
}

func TestExecuteActionsEmptyList(t *testing.T) {
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	var actions []config.Action

	err := ExecuteActions(srcDir, targetDir, actions)
	if err != nil {
		t.Errorf("ExecuteActions with empty list failed: %v", err)
	}
}
