package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateActions_ValidActionsWithExistingFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-validator-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".npmrc"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	actions := []Action{
		{Type: ActionCopy, Path: ".env", Line: 1},
		{Type: ActionSymlink, Path: ".npmrc", Line: 2},
	}

	err = ValidateActions(tmpDir, actions)
	if err != nil {
		t.Errorf("expected no error for valid actions, got: %v", err)
	}
}

func TestValidateActions_MissingSourceFileCopy(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-validator-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	actions := []Action{
		{Type: ActionCopy, Path: "missing.txt", Line: 1},
	}

	err = ValidateActions(tmpDir, actions)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}

	if !strings.Contains(err.Error(), "missing.txt") {
		t.Errorf("expected error to mention missing.txt, got: %v", err)
	}
}

func TestValidateActions_MissingSourceFileSymlink(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-validator-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	actions := []Action{
		{Type: ActionSymlink, Path: "node_modules", Line: 1},
	}

	err = ValidateActions(tmpDir, actions)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}

	if !strings.Contains(err.Error(), "node_modules") {
		t.Errorf("expected error to mention node_modules, got: %v", err)
	}
}

func TestValidateActions_MultipleMissingFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-validator-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	actions := []Action{
		{Type: ActionCopy, Path: "missing1.txt", Line: 1},
		{Type: ActionSymlink, Path: "missing2.txt", Line: 2},
	}

	err = ValidateActions(tmpDir, actions)
	if err == nil {
		t.Fatal("expected error for missing files, got nil")
	}

	// Verify both errors are reported
	if !strings.Contains(err.Error(), "missing1.txt") {
		t.Errorf("expected error to mention missing1.txt, got: %v", err)
	}
	if !strings.Contains(err.Error(), "missing2.txt") {
		t.Errorf("expected error to mention missing2.txt, got: %v", err)
	}
}

func TestValidateActions_RunActionNoSourceCheck(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-validator-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	actions := []Action{
		{Type: ActionRun, Path: "npm install", Line: 1},
		{Type: ActionRun, Path: "echo hello world", Line: 2},
	}

	err = ValidateActions(tmpDir, actions)
	if err != nil {
		t.Errorf("expected no error for run actions, got: %v", err)
	}
}

func TestValidateActions_RunActionEmptyCommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-validator-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	actions := []Action{
		{Type: ActionRun, Path: "", Line: 1},
	}

	err = ValidateActions(tmpDir, actions)
	if err == nil {
		t.Fatal("expected error for empty run command, got nil")
	}
}

func TestValidateActions_UnknownActionType(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-validator-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	actions := []Action{
		{Type: "unknown", Path: "somepath", Line: 1},
	}

	err = ValidateActions(tmpDir, actions)
	if err == nil {
		t.Fatal("expected error for unknown action type, got nil")
	}

	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("expected error to mention unknown type, got: %v", err)
	}
}

func TestValidateActions_MixedValidAndInvalid(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-validator-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create one valid file
	if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	actions := []Action{
		{Type: ActionCopy, Path: ".env", Line: 1},        // valid
		{Type: ActionCopy, Path: "missing.txt", Line: 2}, // invalid
		{Type: ActionRun, Path: "npm install", Line: 3},  // valid
	}

	err = ValidateActions(tmpDir, actions)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}

	// Should only report the invalid one
	if !strings.Contains(err.Error(), "missing.txt") {
		t.Errorf("expected error to mention missing.txt, got: %v", err)
	}

	// Should not mention the valid ones
	if strings.Contains(err.Error(), ".env") {
		t.Errorf("expected error not to mention .env (valid file), got: %v", err)
	}
}

func TestValidateRemoveActions_ValidRunCommands(t *testing.T) {
	actions := []Action{
		{Type: ActionRun, Path: "echo hello"},
		{Type: ActionRun, Path: "git log --oneline"},
	}
	if err := ValidateRemoveActions(actions); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateRemoveActions_RejectsCopyAction(t *testing.T) {
	actions := []Action{
		{Type: ActionCopy, Path: ".env"},
	}
	err := ValidateRemoveActions(actions)
	if err == nil {
		t.Fatal("expected error for copy in remove, got nil")
	}
	if !strings.Contains(err.Error(), "copy actions not allowed") {
		t.Errorf("expected 'copy actions not allowed' error, got: %v", err)
	}
}

func TestValidateRemoveActions_RejectsSymlinkAction(t *testing.T) {
	actions := []Action{
		{Type: ActionSymlink, Path: "node_modules"},
	}
	err := ValidateRemoveActions(actions)
	if err == nil {
		t.Fatal("expected error for symlink in remove, got nil")
	}
	if !strings.Contains(err.Error(), "symlink actions not allowed") {
		t.Errorf("expected 'symlink actions not allowed' error, got: %v", err)
	}
}

func TestValidateRemoveActions_EmptyRunCommand(t *testing.T) {
	actions := []Action{
		{Type: ActionRun, Path: ""},
	}
	err := ValidateRemoveActions(actions)
	if err == nil {
		t.Fatal("expected error for empty run command, got nil")
	}
	if !strings.Contains(err.Error(), "run command cannot be empty") {
		t.Errorf("expected 'run command cannot be empty' error, got: %v", err)
	}
}

func TestValidateRemoveActions_EmptyList(t *testing.T) {
	if err := ValidateRemoveActions([]Action{}); err != nil {
		t.Errorf("expected no error for empty list, got: %v", err)
	}
}

func TestValidateActions_EmptyActionList(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-validator-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	actions := []Action{}

	err = ValidateActions(tmpDir, actions)
	if err != nil {
		t.Errorf("expected no error for empty action list, got: %v", err)
	}
}
