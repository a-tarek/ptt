package setup

import (
	"os"
	"path/filepath"
	"testing"
)

// Copy tests

func TestCopySingleFile(t *testing.T) {
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(srcDir, "test.txt")
	content := []byte("hello world")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Copy file
	targetFile := filepath.Join(targetDir, "test.txt")
	if err := CopyPath(srcFile, targetFile); err != nil {
		t.Fatalf("CopyPath failed: %v", err)
	}

	// Verify content matches
	gotContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("failed to read target file: %v", err)
	}
	if string(gotContent) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", gotContent, content)
	}
}

func TestCopyDirectoryRecursively(t *testing.T) {
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	// Create source directory structure
	subDir := filepath.Join(srcDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	file1 := filepath.Join(srcDir, "file1.txt")
	file2 := filepath.Join(subDir, "file2.txt")

	if err := os.WriteFile(file1, []byte("file1"), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("file2"), 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	// Copy directory
	targetSrc := filepath.Join(targetDir, "srcdir")
	if err := CopyPath(srcDir, targetSrc); err != nil {
		t.Fatalf("CopyPath failed: %v", err)
	}

	// Verify all files exist
	targetFile1 := filepath.Join(targetSrc, "file1.txt")
	targetFile2 := filepath.Join(targetSrc, "subdir", "file2.txt")

	if _, err := os.Stat(targetFile1); err != nil {
		t.Errorf("file1 not copied: %v", err)
	}
	if _, err := os.Stat(targetFile2); err != nil {
		t.Errorf("file2 not copied: %v", err)
	}
}

func TestCopyCreatesParentDirectories(t *testing.T) {
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(srcFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Copy to nested path that doesn't exist
	targetFile := filepath.Join(targetDir, "deep", "nested", "path", "test.txt")
	if err := CopyPath(srcFile, targetFile); err != nil {
		t.Fatalf("CopyPath failed: %v", err)
	}

	// Verify parent directories were created
	if _, err := os.Stat(targetFile); err != nil {
		t.Errorf("target file not created: %v", err)
	}
}

func TestCopyMissingSource(t *testing.T) {
	targetDir := t.TempDir()

	srcFile := filepath.Join(targetDir, "nonexistent.txt")
	targetFile := filepath.Join(targetDir, "target.txt")

	err := CopyPath(srcFile, targetFile)
	if err == nil {
		t.Error("expected error for missing source, got nil")
	}
}

// Symlink tests

func TestSymlinkFile(t *testing.T) {
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(srcFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create symlink
	targetFile := filepath.Join(targetDir, "link.txt")
	if err := CreateSymlink(srcFile, targetFile); err != nil {
		t.Fatalf("CreateSymlink failed: %v", err)
	}

	// Verify it's a symlink and points to correct location
	linkDest, err := os.Readlink(targetFile)
	if err != nil {
		t.Fatalf("failed to read symlink: %v", err)
	}
	if linkDest != srcFile {
		t.Errorf("symlink points to %q, want %q", linkDest, srcFile)
	}
}

func TestSymlinkDirectory(t *testing.T) {
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	// Create source directory with a file
	srcSubDir := filepath.Join(srcDir, "mydir")
	if err := os.MkdirAll(srcSubDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	testFile := filepath.Join(srcSubDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create symlink to directory
	targetLink := filepath.Join(targetDir, "mylink")
	if err := CreateSymlink(srcSubDir, targetLink); err != nil {
		t.Fatalf("CreateSymlink failed: %v", err)
	}

	// Verify symlink exists and points to source dir
	linkDest, err := os.Readlink(targetLink)
	if err != nil {
		t.Fatalf("failed to read symlink: %v", err)
	}
	if linkDest != srcSubDir {
		t.Errorf("symlink points to %q, want %q", linkDest, srcSubDir)
	}
}

func TestSymlinkCreatesParentDirectories(t *testing.T) {
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(srcFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create symlink in nested path that doesn't exist
	targetLink := filepath.Join(targetDir, "deep", "nested", "link.txt")
	if err := CreateSymlink(srcFile, targetLink); err != nil {
		t.Fatalf("CreateSymlink failed: %v", err)
	}

	// Verify parent directories were created and symlink exists
	linkDest, err := os.Readlink(targetLink)
	if err != nil {
		t.Fatalf("failed to read symlink: %v", err)
	}
	if linkDest != srcFile {
		t.Errorf("symlink points to %q, want %q", linkDest, srcFile)
	}
}

func TestSymlinkMissingSource(t *testing.T) {
	targetDir := t.TempDir()

	srcFile := "/nonexistent/source.txt"
	targetLink := filepath.Join(targetDir, "link.txt")

	// Note: os.Symlink doesn't validate source exists, so this will succeed
	// This test documents that behavior - validation should happen upstream
	err := CreateSymlink(srcFile, targetLink)
	if err != nil {
		t.Errorf("CreateSymlink unexpectedly failed: %v", err)
	}
}

// Run tests

func TestRunSimpleCommand(t *testing.T) {
	workDir := t.TempDir()

	err := RunCommand(workDir, "echo hello")
	if err != nil {
		t.Errorf("RunCommand failed: %v", err)
	}
}

func TestRunCommandWithWorkingDirectory(t *testing.T) {
	workDir := t.TempDir()

	// Run command that creates a file
	err := RunCommand(workDir, "echo test > testfile.txt")
	if err != nil {
		t.Fatalf("RunCommand failed: %v", err)
	}

	// Verify file was created in working directory
	testFile := filepath.Join(workDir, "testfile.txt")
	if _, err := os.Stat(testFile); err != nil {
		t.Errorf("file not created in working directory: %v", err)
	}
}

func TestRunFailingCommand(t *testing.T) {
	workDir := t.TempDir()

	err := RunCommand(workDir, "exit 1")
	if err == nil {
		t.Error("expected error for failing command, got nil")
	}

	// Verify error message mentions exit code
	errMsg := err.Error()
	if errMsg != "command failed with exit code 1" {
		t.Errorf("unexpected error message: %q", errMsg)
	}
}

func TestRunNonexistentCommand(t *testing.T) {
	workDir := t.TempDir()

	err := RunCommand(workDir, "nonexistent_command_xyz_12345")
	if err == nil {
		t.Error("expected error for nonexistent command, got nil")
	}
}
