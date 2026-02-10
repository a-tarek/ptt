package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseYAMLFile_BothSections(t *testing.T) {
	dir := t.TempDir()
	content := `
create:
  - copy: .env
  - symlink: node_modules
  - run: npm install

remove:
  - run: echo goodbye
  - run: git log --oneline > /tmp/summary.txt
`
	path := filepath.Join(dir, "default.yml")
	os.WriteFile(path, []byte(content), 0644)

	lc, err := ParseYAMLFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(lc.Create) != 3 {
		t.Fatalf("expected 3 create actions, got %d", len(lc.Create))
	}
	if len(lc.Remove) != 2 {
		t.Fatalf("expected 2 remove actions, got %d", len(lc.Remove))
	}

	if lc.Create[0].Type != ActionCopy || lc.Create[0].Path != ".env" {
		t.Errorf("expected copy .env, got %s %s", lc.Create[0].Type, lc.Create[0].Path)
	}
	if lc.Create[1].Type != ActionSymlink || lc.Create[1].Path != "node_modules" {
		t.Errorf("expected symlink node_modules, got %s %s", lc.Create[1].Type, lc.Create[1].Path)
	}
	if lc.Create[2].Type != ActionRun || lc.Create[2].Path != "npm install" {
		t.Errorf("expected run npm install, got %s %s", lc.Create[2].Type, lc.Create[2].Path)
	}

	if lc.Remove[0].Type != ActionRun || lc.Remove[0].Path != "echo goodbye" {
		t.Errorf("expected run echo goodbye, got %s %s", lc.Remove[0].Type, lc.Remove[0].Path)
	}
}

func TestParseYAMLFile_CreateOnly(t *testing.T) {
	dir := t.TempDir()
	content := `
create:
  - copy: .env
`
	path := filepath.Join(dir, "default.yml")
	os.WriteFile(path, []byte(content), 0644)

	lc, err := ParseYAMLFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(lc.Create) != 1 {
		t.Fatalf("expected 1 create action, got %d", len(lc.Create))
	}
	if lc.Remove != nil {
		t.Errorf("expected nil remove, got %v", lc.Remove)
	}
}

func TestParseYAMLFile_RemoveOnly(t *testing.T) {
	dir := t.TempDir()
	content := `
remove:
  - run: echo bye
`
	path := filepath.Join(dir, "default.yml")
	os.WriteFile(path, []byte(content), 0644)

	lc, err := ParseYAMLFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lc.Create != nil {
		t.Errorf("expected nil create, got %v", lc.Create)
	}
	if len(lc.Remove) != 1 {
		t.Fatalf("expected 1 remove action, got %d", len(lc.Remove))
	}
}

func TestParseYAMLFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "default.yml")
	os.WriteFile(path, []byte(""), 0644)

	lc, err := ParseYAMLFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lc.Create != nil {
		t.Errorf("expected nil create, got %v", lc.Create)
	}
	if lc.Remove != nil {
		t.Errorf("expected nil remove, got %v", lc.Remove)
	}
}

func TestParseYAMLFile_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	content := `
create:
  - [invalid yaml here
`
	path := filepath.Join(dir, "default.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := ParseYAMLFile(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestParseYAMLFile_MultipleFieldsInAction(t *testing.T) {
	dir := t.TempDir()
	content := `
create:
  - copy: .env
    run: npm install
`
	path := filepath.Join(dir, "default.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := ParseYAMLFile(path)
	if err == nil {
		t.Fatal("expected error for action with multiple fields, got nil")
	}
}

func TestParseYAMLFile_NoFieldsInAction(t *testing.T) {
	dir := t.TempDir()
	content := `
create:
  - {}
`
	path := filepath.Join(dir, "default.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := ParseYAMLFile(path)
	if err == nil {
		t.Fatal("expected error for action with no fields, got nil")
	}
}

func TestParseYAMLFile_AllActionTypes(t *testing.T) {
	dir := t.TempDir()
	content := `
create:
  - copy: .env
  - symlink: node_modules
  - run: npm install
`
	path := filepath.Join(dir, "default.yml")
	os.WriteFile(path, []byte(content), 0644)

	lc, err := ParseYAMLFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []struct {
		Type string
		Path string
	}{
		{ActionCopy, ".env"},
		{ActionSymlink, "node_modules"},
		{ActionRun, "npm install"},
	}

	for i, e := range expected {
		if lc.Create[i].Type != e.Type {
			t.Errorf("action[%d]: expected type %s, got %s", i, e.Type, lc.Create[i].Type)
		}
		if lc.Create[i].Path != e.Path {
			t.Errorf("action[%d]: expected path %s, got %s", i, e.Path, lc.Create[i].Path)
		}
	}
}

func TestIsYAMLConfig(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"config.yml", true},
		{"config.yaml", true},
		{"config.YML", true},
		{"config.YAML", true},
		{".pttconfig/default.yml", true},
		{".pttconfig/default", false},
		{".wtconfig", false},
		{"", false},
	}

	for _, tt := range tests {
		got := IsYAMLConfig(tt.path)
		if got != tt.want {
			t.Errorf("IsYAMLConfig(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
