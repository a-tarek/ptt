package config

import (
	"os"
	"strings"
	"testing"
)

func TestParseFile_BasicParsing(t *testing.T) {
	content := `copy .env
symlink node_modules`

	tmpfile, err := os.CreateTemp("", "wtconfig-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	actions, err := ParseFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}

	if actions[0].Type != ActionCopy || actions[0].Path != ".env" {
		t.Errorf("expected copy .env, got %s %s", actions[0].Type, actions[0].Path)
	}

	if actions[1].Type != ActionSymlink || actions[1].Path != "node_modules" {
		t.Errorf("expected symlink node_modules, got %s %s", actions[1].Type, actions[1].Path)
	}
}

func TestParseFile_CommentsSkipped(t *testing.T) {
	content := `# This is a comment
copy .env
# Another comment
symlink node_modules`

	tmpfile, err := os.CreateTemp("", "wtconfig-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	actions, err := ParseFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(actions) != 2 {
		t.Fatalf("expected 2 actions (comments should be skipped), got %d", len(actions))
	}
}

func TestParseFile_BlankLinesSkipped(t *testing.T) {
	content := `copy .env

symlink node_modules

run npm install`

	tmpfile, err := os.CreateTemp("", "wtconfig-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	actions, err := ParseFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(actions) != 3 {
		t.Fatalf("expected 3 actions (blank lines should be skipped), got %d", len(actions))
	}
}

func TestParseFile_RunCommandWithSpaces(t *testing.T) {
	content := `run npm install --save-dev typescript`

	tmpfile, err := os.CreateTemp("", "wtconfig-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	actions, err := ParseFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}

	if actions[0].Type != ActionRun {
		t.Errorf("expected run action, got %s", actions[0].Type)
	}

	if actions[0].Path != "npm install --save-dev typescript" {
		t.Errorf("expected full command, got: %s", actions[0].Path)
	}
}

func TestParseFile_TrailingWhitespace(t *testing.T) {
	content := "copy .env  \t\n"

	tmpfile, err := os.CreateTemp("", "wtconfig-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	actions, err := ParseFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}

	if actions[0].Path != ".env" {
		t.Errorf("expected path to be trimmed to .env, got: '%s'", actions[0].Path)
	}
}

func TestParseFile_LeadingWhitespace(t *testing.T) {
	content := `  copy .env
	symlink node_modules`

	tmpfile, err := os.CreateTemp("", "wtconfig-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	actions, err := ParseFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}

	if actions[0].Type != ActionCopy || actions[0].Path != ".env" {
		t.Errorf("expected copy .env, got %s %s", actions[0].Type, actions[0].Path)
	}
}

func TestParseFile_InvalidLineNoSpace(t *testing.T) {
	content := `copy .env
badline
symlink node_modules`

	tmpfile, err := os.CreateTemp("", "wtconfig-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	_, err = ParseFile(tmpfile.Name())
	if err == nil {
		t.Fatal("expected error for invalid line, got nil")
	}

	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("expected error to mention line 2, got: %v", err)
	}
}

func TestParseFile_LineNumbersInErrors(t *testing.T) {
	content := `copy .env
symlink node_modules
run npm install

invalidline
copy .npmrc`

	tmpfile, err := os.CreateTemp("", "wtconfig-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	_, err = ParseFile(tmpfile.Name())
	if err == nil {
		t.Fatal("expected error for invalid line, got nil")
	}

	if !strings.Contains(err.Error(), "line 5") {
		t.Errorf("expected error to mention line 5, got: %v", err)
	}
}

func TestParseFile_MixedContent(t *testing.T) {
	content := `# Comment at start

copy .env
# Another comment
symlink node_modules

run npm install
  # Indented comment

copy .npmrc`

	tmpfile, err := os.CreateTemp("", "wtconfig-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	actions, err := ParseFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(actions) != 4 {
		t.Fatalf("expected 4 actions, got %d", len(actions))
	}

	expectedActions := []struct {
		typ  string
		path string
	}{
		{ActionCopy, ".env"},
		{ActionSymlink, "node_modules"},
		{ActionRun, "npm install"},
		{ActionCopy, ".npmrc"},
	}

	for i, expected := range expectedActions {
		if actions[i].Type != expected.typ || actions[i].Path != expected.path {
			t.Errorf("action %d: expected %s %s, got %s %s",
				i, expected.typ, expected.path, actions[i].Type, actions[i].Path)
		}
	}
}

func TestParseFile_EmptyFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "wtconfig-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	actions, err := ParseFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ParseFile failed on empty file: %v", err)
	}

	if len(actions) != 0 {
		t.Fatalf("expected empty action slice for empty file, got %d actions", len(actions))
	}
}
