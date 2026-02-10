package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveConfigPath_BareName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-resolve-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config file
	configPath := filepath.Join(tmpDir, ".wtconfig-ci")
	if err := os.WriteFile(configPath, []byte("copy .env"), 0644); err != nil {
		t.Fatal(err)
	}

	resolved, err := ResolveConfigPath(tmpDir, "ci")
	if err != nil {
		t.Fatalf("ResolveConfigPath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".wtconfig-ci")
	if resolved != expected {
		t.Errorf("expected %s, got %s", expected, resolved)
	}
}

func TestResolveConfigPath_BareNameExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-resolve-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config file
	configPath := filepath.Join(tmpDir, ".wtconfig-staging")
	if err := os.WriteFile(configPath, []byte("copy .env"), 0644); err != nil {
		t.Fatal(err)
	}

	resolved, err := ResolveConfigPath(tmpDir, "staging")
	if err != nil {
		t.Fatalf("ResolveConfigPath failed: %v", err)
	}

	if resolved != configPath {
		t.Errorf("expected %s, got %s", configPath, resolved)
	}
}

func TestResolveConfigPath_BareNameNonexistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-resolve-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = ResolveConfigPath(tmpDir, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent config, got nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to mention 'not found', got: %v", err)
	}
}

func TestResolveConfigPath_ExactPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-resolve-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create nested directory and config
	configDir := filepath.Join(tmpDir, "configs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "my-config")
	if err := os.WriteFile(configPath, []byte("copy .env"), 0644); err != nil {
		t.Fatal(err)
	}

	// Path contains "/" so treated as exact path (absolute path)
	resolved, err := ResolveConfigPath(tmpDir, configPath)
	if err != nil {
		t.Fatalf("ResolveConfigPath failed: %v", err)
	}

	if resolved != configPath {
		t.Errorf("expected %s, got %s", configPath, resolved)
	}
}

func TestResolveConfigPath_ExactPathExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-resolve-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config at exact path
	configPath := filepath.Join(tmpDir, "custom-config")
	if err := os.WriteFile(configPath, []byte("copy .env"), 0644); err != nil {
		t.Fatal(err)
	}

	// Use path relative to tmpDir (working as if tmpDir is cwd)
	resolved, err := ResolveConfigPath("", filepath.Join(tmpDir, "custom-config"))
	if err != nil {
		t.Fatalf("ResolveConfigPath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "custom-config")
	if resolved != expected {
		t.Errorf("expected %s, got %s", expected, resolved)
	}
}

func TestResolveConfigPath_ExactPathNonexistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-resolve-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = ResolveConfigPath(tmpDir, "configs/nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent config, got nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to mention 'not found', got: %v", err)
	}
}

func TestResolveConfigPath_DefaultConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-resolve-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create default config file
	configPath := filepath.Join(tmpDir, ".wtconfig")
	if err := os.WriteFile(configPath, []byte("copy .env"), 0644); err != nil {
		t.Fatal(err)
	}

	resolved, err := ResolveConfigPath(tmpDir, "")
	if err != nil {
		t.Fatalf("ResolveConfigPath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".wtconfig")
	if resolved != expected {
		t.Errorf("expected %s, got %s", expected, resolved)
	}
}

func TestResolveConfigPath_PttconfigDefault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-resolve-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .pttconfig/default
	pttconfigDir := filepath.Join(tmpDir, ".pttconfig")
	if err := os.MkdirAll(pttconfigDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(pttconfigDir, "default")
	if err := os.WriteFile(configPath, []byte("copy .env"), 0644); err != nil {
		t.Fatal(err)
	}

	resolved, err := ResolveConfigPath(tmpDir, "")
	if err != nil {
		t.Fatalf("ResolveConfigPath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".pttconfig", "default")
	if resolved != expected {
		t.Errorf("expected %s, got %s", expected, resolved)
	}
}

func TestResolveConfigPath_PttconfigNamed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-resolve-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .pttconfig/ci
	pttconfigDir := filepath.Join(tmpDir, ".pttconfig")
	if err := os.MkdirAll(pttconfigDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(pttconfigDir, "ci")
	if err := os.WriteFile(configPath, []byte("copy .env"), 0644); err != nil {
		t.Fatal(err)
	}

	resolved, err := ResolveConfigPath(tmpDir, "ci")
	if err != nil {
		t.Fatalf("ResolveConfigPath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".pttconfig", "ci")
	if resolved != expected {
		t.Errorf("expected %s, got %s", expected, resolved)
	}
}

func TestResolveConfigPath_FallbackToWtconfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-resolve-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create only .wtconfig (no .pttconfig directory)
	configPath := filepath.Join(tmpDir, ".wtconfig")
	if err := os.WriteFile(configPath, []byte("copy .env"), 0644); err != nil {
		t.Fatal(err)
	}

	resolved, err := ResolveConfigPath(tmpDir, "")
	if err != nil {
		t.Fatalf("ResolveConfigPath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".wtconfig")
	if resolved != expected {
		t.Errorf("expected %s, got %s", expected, resolved)
	}
}

func TestResolveConfigPath_FallbackToWtconfigNamed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-resolve-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create only .wtconfig-ci (no .pttconfig directory)
	configPath := filepath.Join(tmpDir, ".wtconfig-ci")
	if err := os.WriteFile(configPath, []byte("copy .env"), 0644); err != nil {
		t.Fatal(err)
	}

	resolved, err := ResolveConfigPath(tmpDir, "ci")
	if err != nil {
		t.Fatalf("ResolveConfigPath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".wtconfig-ci")
	if resolved != expected {
		t.Errorf("expected %s, got %s", expected, resolved)
	}
}

func TestResolveConfigPath_YAMLDefaultPriority(t *testing.T) {
	tmpDir := t.TempDir()

	// Create both .pttconfig/default.yml and .pttconfig/default (text)
	pttconfigDir := filepath.Join(tmpDir, ".pttconfig")
	os.MkdirAll(pttconfigDir, 0755)
	os.WriteFile(filepath.Join(pttconfigDir, "default.yml"), []byte("create:\n  - copy: .env\n"), 0644)
	os.WriteFile(filepath.Join(pttconfigDir, "default"), []byte("copy .env\n"), 0644)

	resolved, err := ResolveConfigPath(tmpDir, "")
	if err != nil {
		t.Fatalf("ResolveConfigPath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".pttconfig", "default.yml")
	if resolved != expected {
		t.Errorf("expected %s, got %s", expected, resolved)
	}
}

func TestResolveConfigPath_YAMLNamedConfig(t *testing.T) {
	tmpDir := t.TempDir()

	pttconfigDir := filepath.Join(tmpDir, ".pttconfig")
	os.MkdirAll(pttconfigDir, 0755)
	os.WriteFile(filepath.Join(pttconfigDir, "ci.yml"), []byte("create:\n  - run: npm ci\n"), 0644)

	resolved, err := ResolveConfigPath(tmpDir, "ci")
	if err != nil {
		t.Fatalf("ResolveConfigPath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".pttconfig", "ci.yml")
	if resolved != expected {
		t.Errorf("expected %s, got %s", expected, resolved)
	}
}

func TestResolveConfigPath_YAMLFallbackToText(t *testing.T) {
	tmpDir := t.TempDir()

	// Only text config exists, no YAML
	pttconfigDir := filepath.Join(tmpDir, ".pttconfig")
	os.MkdirAll(pttconfigDir, 0755)
	os.WriteFile(filepath.Join(pttconfigDir, "default"), []byte("copy .env\n"), 0644)

	resolved, err := ResolveConfigPath(tmpDir, "")
	if err != nil {
		t.Fatalf("ResolveConfigPath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".pttconfig", "default")
	if resolved != expected {
		t.Errorf("expected %s, got %s", expected, resolved)
	}
}

func TestResolveConfigPath_YAMLAlternateExtension(t *testing.T) {
	tmpDir := t.TempDir()

	pttconfigDir := filepath.Join(tmpDir, ".pttconfig")
	os.MkdirAll(pttconfigDir, 0755)
	os.WriteFile(filepath.Join(pttconfigDir, "default.yaml"), []byte("create:\n  - copy: .env\n"), 0644)

	resolved, err := ResolveConfigPath(tmpDir, "")
	if err != nil {
		t.Fatalf("ResolveConfigPath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".pttconfig", "default.yaml")
	if resolved != expected {
		t.Errorf("expected %s, got %s", expected, resolved)
	}
}

func TestResolveConfigPath_PttconfigTakesPriority(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wt-resolve-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create both .pttconfig/default and .wtconfig
	pttconfigDir := filepath.Join(tmpDir, ".pttconfig")
	if err := os.MkdirAll(pttconfigDir, 0755); err != nil {
		t.Fatal(err)
	}

	pttConfigPath := filepath.Join(pttconfigDir, "default")
	if err := os.WriteFile(pttConfigPath, []byte("# pttconfig"), 0644); err != nil {
		t.Fatal(err)
	}

	wtConfigPath := filepath.Join(tmpDir, ".wtconfig")
	if err := os.WriteFile(wtConfigPath, []byte("# wtconfig"), 0644); err != nil {
		t.Fatal(err)
	}

	resolved, err := ResolveConfigPath(tmpDir, "")
	if err != nil {
		t.Fatalf("ResolveConfigPath failed: %v", err)
	}

	// Should resolve to .pttconfig/default, not .wtconfig
	expected := filepath.Join(tmpDir, ".pttconfig", "default")
	if resolved != expected {
		t.Errorf("expected %s, got %s (priority not working)", expected, resolved)
	}
}
