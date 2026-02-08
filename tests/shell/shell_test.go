package shell_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var (
	wtBinaryPath string
	buildOnce    sync.Once
)

// buildWtBinary compiles the wt binary once and returns its path.
// Uses sync.Once to ensure we only build once per test run.
func buildWtBinary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		tmpDir, err := os.MkdirTemp("", "wt-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir for binary: %v", err)
		}
		// Note: Don't clean up tmpDir - it's used by all tests

		wtBinaryPath = filepath.Join(tmpDir, "wt")
		projectRoot := filepath.Join("..", "..")

		cmd := exec.Command("go", "build", "-o", wtBinaryPath, ".")
		cmd.Dir = projectRoot
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to build wt binary: %v\nOutput: %s", err, output)
		}
	})
	return wtBinaryPath
}

// runSetup executes the setup.sh script to create git fixtures.
func runSetup(t *testing.T, tmpDir string) {
	t.Helper()
	setupScript := filepath.Join("testdata", "setup.sh")
	cmd := exec.Command("/usr/bin/env", "bash", setupScript, tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run setup.sh: %v\nOutput: %s", err, output)
	}
}

// wrapperPath returns the absolute path to the wrapper script for the given shell.
func wrapperPath(t *testing.T, shell string) string {
	t.Helper()
	projectRoot := filepath.Join("..", "..")
	absPath, err := filepath.Abs(filepath.Join(projectRoot, "internal", "shell", "templates", fmt.Sprintf("wrapper.%s", shell)))
	if err != nil {
		t.Fatalf("Failed to resolve wrapper path: %v", err)
	}
	return absPath
}

// parsePWD extracts the PWD from command output in the format "RESULT_PWD=/path/to/dir".
func parsePWD(output string) (string, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "RESULT_PWD=") {
			return strings.TrimPrefix(line, "RESULT_PWD="), nil
		}
	}
	return "", fmt.Errorf("RESULT_PWD not found in output: %s", output)
}

// TestBashWrapperGoto tests that the bash wrapper successfully changes directory with 'wt goto'.
func TestBashWrapperGoto(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	bashPath, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not available")
	}

	tmpBin := buildWtBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		export PATH="%s:$PATH"
		source "%s"
		cd "%s/main"
		wt goto feature
		echo "RESULT_PWD=$PWD"
	`, filepath.Dir(tmpBin), wrapperPath(t, "bash"), tmpDir)

	cmd := exec.Command(bashPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	pwd, err := parsePWD(string(output))
	if err != nil {
		t.Fatal(err)
	}

	expectedSuffix := "/feature"
	if !strings.HasSuffix(pwd, expectedSuffix) {
		t.Errorf("Expected PWD to end with %q, got: %s", expectedSuffix, pwd)
	}
}

// TestBashWrapperHome tests that the bash wrapper successfully changes directory with 'wt home'.
func TestBashWrapperHome(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	bashPath, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not available")
	}

	tmpBin := buildWtBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		export PATH="%s:$PATH"
		source "%s"
		cd "%s/feature"
		wt home
		echo "RESULT_PWD=$PWD"
	`, filepath.Dir(tmpBin), wrapperPath(t, "bash"), tmpDir)

	cmd := exec.Command(bashPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	pwd, err := parsePWD(string(output))
	if err != nil {
		t.Fatal(err)
	}

	expectedSuffix := "/main"
	if !strings.HasSuffix(pwd, expectedSuffix) {
		t.Errorf("Expected PWD to end with %q, got: %s", expectedSuffix, pwd)
	}
}

// TestBashWrapperNew tests that the bash wrapper successfully changes directory with 'wt new'.
func TestBashWrapperNew(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	bashPath, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not available")
	}

	tmpBin := buildWtBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		export PATH="%s:$PATH"
		source "%s"
		cd "%s/main"
		wt new test-branch
		echo "RESULT_PWD=$PWD"
	`, filepath.Dir(tmpBin), wrapperPath(t, "bash"), tmpDir)

	cmd := exec.Command(bashPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	pwd, err := parsePWD(string(output))
	if err != nil {
		t.Fatal(err)
	}

	expectedSuffix := "/test-branch"
	if !strings.HasSuffix(pwd, expectedSuffix) {
		t.Errorf("Expected PWD to end with %q, got: %s", expectedSuffix, pwd)
	}
}

// TestZshWrapperGoto tests that the zsh wrapper successfully changes directory with 'wt goto'.
func TestZshWrapperGoto(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	zshPath, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not available")
	}

	tmpBin := buildWtBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		export PATH="%s:$PATH"
		source "%s"
		cd "%s/main"
		wt goto feature
		echo "RESULT_PWD=$PWD"
	`, filepath.Dir(tmpBin), wrapperPath(t, "zsh"), tmpDir)

	cmd := exec.Command(zshPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	pwd, err := parsePWD(string(output))
	if err != nil {
		t.Fatal(err)
	}

	expectedSuffix := "/feature"
	if !strings.HasSuffix(pwd, expectedSuffix) {
		t.Errorf("Expected PWD to end with %q, got: %s", expectedSuffix, pwd)
	}
}

// TestZshWrapperHome tests that the zsh wrapper successfully changes directory with 'wt home'.
func TestZshWrapperHome(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	zshPath, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not available")
	}

	tmpBin := buildWtBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		export PATH="%s:$PATH"
		source "%s"
		cd "%s/feature"
		wt home
		echo "RESULT_PWD=$PWD"
	`, filepath.Dir(tmpBin), wrapperPath(t, "zsh"), tmpDir)

	cmd := exec.Command(zshPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	pwd, err := parsePWD(string(output))
	if err != nil {
		t.Fatal(err)
	}

	expectedSuffix := "/main"
	if !strings.HasSuffix(pwd, expectedSuffix) {
		t.Errorf("Expected PWD to end with %q, got: %s", expectedSuffix, pwd)
	}
}

// TestZshWrapperNew tests that the zsh wrapper successfully changes directory with 'wt new'.
func TestZshWrapperNew(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	zshPath, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not available")
	}

	tmpBin := buildWtBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		export PATH="%s:$PATH"
		source "%s"
		cd "%s/main"
		wt new test-branch
		echo "RESULT_PWD=$PWD"
	`, filepath.Dir(tmpBin), wrapperPath(t, "zsh"), tmpDir)

	cmd := exec.Command(zshPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	pwd, err := parsePWD(string(output))
	if err != nil {
		t.Fatal(err)
	}

	expectedSuffix := "/test-branch"
	if !strings.HasSuffix(pwd, expectedSuffix) {
		t.Errorf("Expected PWD to end with %q, got: %s", expectedSuffix, pwd)
	}
}

// TestFishWrapperGoto tests that the fish wrapper successfully changes directory with 'wt goto'.
func TestFishWrapperGoto(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	fishPath, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish not available")
	}

	tmpBin := buildWtBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		set -x PATH "%s" $PATH
		source "%s"
		cd "%s/main"
		wt goto feature
		echo "RESULT_PWD=$PWD"
	`, filepath.Dir(tmpBin), wrapperPath(t, "fish"), tmpDir)

	cmd := exec.Command(fishPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	pwd, err := parsePWD(string(output))
	if err != nil {
		t.Fatal(err)
	}

	expectedSuffix := "/feature"
	if !strings.HasSuffix(pwd, expectedSuffix) {
		t.Errorf("Expected PWD to end with %q, got: %s", expectedSuffix, pwd)
	}
}

// TestFishWrapperHome tests that the fish wrapper successfully changes directory with 'wt home'.
func TestFishWrapperHome(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	fishPath, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish not available")
	}

	tmpBin := buildWtBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		set -x PATH "%s" $PATH
		source "%s"
		cd "%s/feature"
		wt home
		echo "RESULT_PWD=$PWD"
	`, filepath.Dir(tmpBin), wrapperPath(t, "fish"), tmpDir)

	cmd := exec.Command(fishPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	pwd, err := parsePWD(string(output))
	if err != nil {
		t.Fatal(err)
	}

	expectedSuffix := "/main"
	if !strings.HasSuffix(pwd, expectedSuffix) {
		t.Errorf("Expected PWD to end with %q, got: %s", expectedSuffix, pwd)
	}
}

// TestFishWrapperNew tests that the fish wrapper successfully changes directory with 'wt new'.
func TestFishWrapperNew(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	fishPath, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish not available")
	}

	tmpBin := buildWtBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		set -x PATH "%s" $PATH
		source "%s"
		cd "%s/main"
		wt new test-branch
		echo "RESULT_PWD=$PWD"
	`, filepath.Dir(tmpBin), wrapperPath(t, "fish"), tmpDir)

	cmd := exec.Command(fishPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	pwd, err := parsePWD(string(output))
	if err != nil {
		t.Fatal(err)
	}

	expectedSuffix := "/test-branch"
	if !strings.HasSuffix(pwd, expectedSuffix) {
		t.Errorf("Expected PWD to end with %q, got: %s", expectedSuffix, pwd)
	}
}
