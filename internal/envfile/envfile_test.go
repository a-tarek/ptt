package envfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseKeyValue(t *testing.T) {
	path := writeTempEnv(t, "PORT=3000\nHOST=localhost\n")
	f, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(f.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(f.Entries))
	}
	if f.Entries[0].Key != "PORT" || f.Entries[0].Value != "3000" {
		t.Errorf("entry 0: got %q=%q", f.Entries[0].Key, f.Entries[0].Value)
	}
	if f.Entries[1].Key != "HOST" || f.Entries[1].Value != "localhost" {
		t.Errorf("entry 1: got %q=%q", f.Entries[1].Key, f.Entries[1].Value)
	}
}

func TestParseDoubleQuoted(t *testing.T) {
	path := writeTempEnv(t, `DB_URL="postgres://localhost/db"`+"\n")
	f, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	e := f.Entries[0]
	if e.Key != "DB_URL" || e.Value != "postgres://localhost/db" {
		t.Errorf("got %q=%q", e.Key, e.Value)
	}
	if e.QuoteChar != '"' {
		t.Errorf("expected double quote, got %c", e.QuoteChar)
	}
}

func TestParseSingleQuoted(t *testing.T) {
	path := writeTempEnv(t, "SECRET='my secret'\n")
	f, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	e := f.Entries[0]
	if e.Key != "SECRET" || e.Value != "my secret" {
		t.Errorf("got %q=%q", e.Key, e.Value)
	}
	if e.QuoteChar != '\'' {
		t.Errorf("expected single quote, got %c", e.QuoteChar)
	}
}

func TestParseExportPrefix(t *testing.T) {
	path := writeTempEnv(t, "export PORT=3000\n")
	f, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	e := f.Entries[0]
	if e.Key != "PORT" || e.Value != "3000" {
		t.Errorf("got %q=%q", e.Key, e.Value)
	}
	if !e.HasExport {
		t.Error("expected HasExport=true")
	}
}

func TestParseCommentsAndBlankLines(t *testing.T) {
	content := "# this is a comment\n\nPORT=3000\n# another comment\nHOST=localhost\n"
	path := writeTempEnv(t, content)
	f, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(f.Entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(f.Entries))
	}
	// Comments and blank lines should have empty Key
	if f.Entries[0].Key != "" {
		t.Error("comment line should have empty key")
	}
	if f.Entries[1].Key != "" {
		t.Error("blank line should have empty key")
	}
	if f.Entries[2].Key != "PORT" {
		t.Error("expected PORT")
	}
	if f.Entries[3].Key != "" {
		t.Error("second comment should have empty key")
	}
	if f.Entries[4].Key != "HOST" {
		t.Error("expected HOST")
	}
}

func TestGet(t *testing.T) {
	path := writeTempEnv(t, "PORT=3000\nHOST=localhost\n")
	f, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	val, ok := f.Get("PORT")
	if !ok || val != "3000" {
		t.Errorf("Get PORT: got %q, %v", val, ok)
	}

	_, ok = f.Get("MISSING")
	if ok {
		t.Error("Get MISSING: expected not found")
	}
}

func TestSet(t *testing.T) {
	path := writeTempEnv(t, "PORT=3000\nHOST=localhost\n")
	f, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if err := f.Set("PORT", "4000"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, _ := f.Get("PORT")
	if val != "4000" {
		t.Errorf("after Set: got %q, want 4000", val)
	}
}

func TestSetPreservesQuotes(t *testing.T) {
	path := writeTempEnv(t, `DB_URL="postgres://localhost/db"`+"\n")
	f, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if err := f.Set("DB_URL", "postgres://localhost/other"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if f.Entries[0].RawLine != `DB_URL="postgres://localhost/other"` {
		t.Errorf("expected quotes preserved, got: %s", f.Entries[0].RawLine)
	}
}

func TestSetPreservesExport(t *testing.T) {
	path := writeTempEnv(t, "export PORT=3000\n")
	f, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if err := f.Set("PORT", "4000"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if f.Entries[0].RawLine != "export PORT=4000" {
		t.Errorf("expected export preserved, got: %s", f.Entries[0].RawLine)
	}
}

func TestSetMissingKey(t *testing.T) {
	path := writeTempEnv(t, "PORT=3000\n")
	f, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	err = f.Set("MISSING", "value")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "MISSING") {
		t.Errorf("error should mention key, got: %v", err)
	}
}

func TestWriteRoundTrip(t *testing.T) {
	content := "# Config\nexport PORT=3000\nHOST=localhost\nDB_URL=\"postgres://localhost/db\"\nSECRET='mysecret'\n\n# End\n"
	path := writeTempEnv(t, content)
	f, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	outPath := filepath.Join(t.TempDir(), "out.env")
	if err := f.Write(outPath); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output failed: %v", err)
	}
	if string(got) != content {
		t.Errorf("round-trip mismatch:\ngot:\n%s\nwant:\n%s", got, content)
	}
}

func TestWriteAfterSet(t *testing.T) {
	content := "PORT=3000\nHOST=localhost\n"
	path := writeTempEnv(t, content)
	f, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if err := f.Set("PORT", "4000"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	outPath := filepath.Join(t.TempDir(), "out.env")
	if err := f.Write(outPath); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output failed: %v", err)
	}
	expected := "PORT=4000\nHOST=localhost\n"
	if string(got) != expected {
		t.Errorf("got:\n%s\nwant:\n%s", got, expected)
	}
}

func writeTempEnv(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp env: %v", err)
	}
	return path
}
