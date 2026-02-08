package config

import (
	"strings"
	"testing"
)

func TestBuildActionsFromFlags_SingleCopy(t *testing.T) {
	copyPaths := []string{".env"}
	symlinkPaths := []string{}
	runCommands := []string{}
	osArgs := []string{"wt", "new", "test", "--copy", ".env"}

	actions := BuildActionsFromFlags(copyPaths, symlinkPaths, runCommands, osArgs)

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}

	if actions[0].Type != ActionCopy || actions[0].Path != ".env" {
		t.Errorf("expected copy .env, got %s %s", actions[0].Type, actions[0].Path)
	}
}

func TestBuildActionsFromFlags_SingleSymlink(t *testing.T) {
	copyPaths := []string{}
	symlinkPaths := []string{"node_modules"}
	runCommands := []string{}
	osArgs := []string{"wt", "new", "test", "--symlink", "node_modules"}

	actions := BuildActionsFromFlags(copyPaths, symlinkPaths, runCommands, osArgs)

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}

	if actions[0].Type != ActionSymlink || actions[0].Path != "node_modules" {
		t.Errorf("expected symlink node_modules, got %s %s", actions[0].Type, actions[0].Path)
	}
}

func TestBuildActionsFromFlags_SingleRun(t *testing.T) {
	copyPaths := []string{}
	symlinkPaths := []string{}
	runCommands := []string{"npm install"}
	osArgs := []string{"wt", "new", "test", "--run", "npm install"}

	actions := BuildActionsFromFlags(copyPaths, symlinkPaths, runCommands, osArgs)

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}

	if actions[0].Type != ActionRun || actions[0].Path != "npm install" {
		t.Errorf("expected run 'npm install', got %s %s", actions[0].Type, actions[0].Path)
	}
}

func TestBuildActionsFromFlags_PreservesCliOrder(t *testing.T) {
	copyPaths := []string{".env"}
	symlinkPaths := []string{"node_modules"}
	runCommands := []string{"npm install"}
	// CLI order: run, copy, symlink
	osArgs := []string{"wt", "new", "test", "--run", "npm install", "--copy", ".env", "--symlink", "node_modules"}

	actions := BuildActionsFromFlags(copyPaths, symlinkPaths, runCommands, osArgs)

	if len(actions) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(actions))
	}

	// Order should match CLI: run, copy, symlink
	if actions[0].Type != ActionRun || actions[0].Path != "npm install" {
		t.Errorf("expected run 'npm install' first, got %s %s", actions[0].Type, actions[0].Path)
	}

	if actions[1].Type != ActionCopy || actions[1].Path != ".env" {
		t.Errorf("expected copy .env second, got %s %s", actions[1].Type, actions[1].Path)
	}

	if actions[2].Type != ActionSymlink || actions[2].Path != "node_modules" {
		t.Errorf("expected symlink node_modules third, got %s %s", actions[2].Type, actions[2].Path)
	}
}

func TestBuildActionsFromFlags_MultipleOfSameType(t *testing.T) {
	copyPaths := []string{".env", ".npmrc"}
	symlinkPaths := []string{}
	runCommands := []string{}
	osArgs := []string{"wt", "new", "test", "--copy", ".env", "--copy", ".npmrc"}

	actions := BuildActionsFromFlags(copyPaths, symlinkPaths, runCommands, osArgs)

	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}

	if actions[0].Type != ActionCopy || actions[0].Path != ".env" {
		t.Errorf("expected copy .env, got %s %s", actions[0].Type, actions[0].Path)
	}

	if actions[1].Type != ActionCopy || actions[1].Path != ".npmrc" {
		t.Errorf("expected copy .npmrc, got %s %s", actions[1].Type, actions[1].Path)
	}
}

func TestBuildActionsFromFlags_EmptyInputs(t *testing.T) {
	copyPaths := []string{}
	symlinkPaths := []string{}
	runCommands := []string{}

	actions := BuildActionsFromFlags(copyPaths, symlinkPaths, runCommands, nil)

	if len(actions) != 0 {
		t.Fatalf("expected empty action slice, got %d actions", len(actions))
	}
}

func TestBuildActionsFromFlags_EqualsStyle(t *testing.T) {
	copyPaths := []string{".env"}
	runCommands := []string{"npm install"}
	// Using --flag=value style
	osArgs := []string{"wt", "new", "test", "--run=npm install", "--copy=.env"}

	actions := BuildActionsFromFlags(copyPaths, []string{}, runCommands, osArgs)

	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}

	if actions[0].Type != ActionRun || actions[0].Path != "npm install" {
		t.Errorf("expected run first, got %s %s", actions[0].Type, actions[0].Path)
	}

	if actions[1].Type != ActionCopy || actions[1].Path != ".env" {
		t.Errorf("expected copy second, got %s %s", actions[1].Type, actions[1].Path)
	}
}

func TestCheckDuplicatePaths_Duplicate(t *testing.T) {
	copyPaths := []string{".env"}
	symlinkPaths := []string{".env"}

	err := CheckDuplicatePaths(copyPaths, symlinkPaths)
	if err == nil {
		t.Fatal("expected error for duplicate path, got nil")
	}

	if !strings.Contains(err.Error(), ".env") {
		t.Errorf("expected error to mention .env, got: %v", err)
	}
}

func TestCheckDuplicatePaths_DuplicateSameType(t *testing.T) {
	copyPaths := []string{".env", ".env"}
	symlinkPaths := []string{}

	err := CheckDuplicatePaths(copyPaths, symlinkPaths)
	if err == nil {
		t.Fatal("expected error for duplicate path, got nil")
	}

	if !strings.Contains(err.Error(), ".env") {
		t.Errorf("expected error to mention .env, got: %v", err)
	}
}

func TestCheckDuplicatePaths_NoDuplicate(t *testing.T) {
	copyPaths := []string{".env"}
	symlinkPaths := []string{"node_modules"}

	err := CheckDuplicatePaths(copyPaths, symlinkPaths)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}
