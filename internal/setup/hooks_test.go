package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/a-tarek/ptt/internal/config"
)

func TestRunPreRemoveHooks_AllSucceed(t *testing.T) {
	dir := t.TempDir()

	actions := []config.Action{
		{Type: config.ActionRun, Path: "echo hello"},
		{Type: config.ActionRun, Path: "echo world"},
	}

	err := RunPreRemoveHooks(dir, actions, nil, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPreRemoveHooks_FirstFails(t *testing.T) {
	dir := t.TempDir()

	actions := []config.Action{
		{Type: config.ActionRun, Path: "exit 1"},
		{Type: config.ActionRun, Path: "echo should-not-run"},
	}

	err := RunPreRemoveHooks(dir, actions, nil, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunPreRemoveHooks_MiddleFails(t *testing.T) {
	dir := t.TempDir()
	markerPath := filepath.Join(dir, "third-ran")

	actions := []config.Action{
		{Type: config.ActionRun, Path: "echo first"},
		{Type: config.ActionRun, Path: "exit 1"},
		{Type: config.ActionRun, Path: "touch " + markerPath},
	}

	err := RunPreRemoveHooks(dir, actions, nil, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Third hook should not have run
	if _, err := os.Stat(markerPath); err == nil {
		t.Error("third hook should not have run after middle hook failed")
	}
}

func TestRunPreRemoveHooks_EmptyList(t *testing.T) {
	dir := t.TempDir()

	err := RunPreRemoveHooks(dir, []config.Action{}, nil, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPreRemoveHooks_WritesMarkerFile(t *testing.T) {
	dir := t.TempDir()
	markerPath := filepath.Join(dir, "hook-marker")

	actions := []config.Action{
		{Type: config.ActionRun, Path: "touch " + markerPath},
	}

	err := RunPreRemoveHooks(dir, actions, nil, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Error("marker file should exist after hook execution")
	}
}
