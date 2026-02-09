package shell_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"os"
)

var (
	pttBinaryPath string
	buildOnce     sync.Once
)

// buildPttBinary compiles the ptt binary once and returns its path.
// Uses sync.Once to ensure we only build once per test run.
func buildPttBinary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		tmpDir, err := os.MkdirTemp("", "ptt-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir for binary: %v", err)
		}
		// Note: Don't clean up tmpDir - it's used by all tests

		pttBinaryPath = filepath.Join(tmpDir, "ptt")
		projectRoot := filepath.Join("..", "..")

		cmd := exec.Command("go", "build", "-o", pttBinaryPath, ".")
		cmd.Dir = projectRoot
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to build ptt binary: %v\nOutput: %s", err, output)
		}
	})
	return pttBinaryPath
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

// TestBashWrapperCd tests that the bash wrapper successfully changes directory with 'ptt cd'.
func TestBashWrapperCd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	bashPath, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not available")
	}

	tmpBin := buildPttBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		export SHELL="%s"
		eval "$('%s' shell-init)"
		cd "%s/main"
		ptt cd feature
		echo "RESULT_PWD=$PWD"
	`, bashPath, tmpBin, tmpDir)

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

// TestBashWrapperCdHome tests that the bash wrapper successfully navigates home with 'ptt cd' (no args).
func TestBashWrapperCdHome(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	bashPath, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not available")
	}

	tmpBin := buildPttBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		export SHELL="%s"
		eval "$('%s' shell-init)"
		cd "%s/feature"
		ptt cd
		echo "RESULT_PWD=$PWD"
	`, bashPath, tmpBin, tmpDir)

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

// TestBashWrapperNew tests that the bash wrapper successfully changes directory with 'ptt new'.
func TestBashWrapperNew(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	bashPath, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not available")
	}

	tmpBin := buildPttBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		export SHELL="%s"
		eval "$('%s' shell-init)"
		cd "%s/main"
		ptt new test-branch
		echo "RESULT_PWD=$PWD"
	`, bashPath, tmpBin, tmpDir)

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

// TestBashWrapperPassthrough tests that non-cd commands pass through correctly.
func TestBashWrapperPassthrough(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	bashPath, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not available")
	}

	tmpBin := buildPttBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		export SHELL="%s"
		eval "$('%s' shell-init)"
		cd "%s/main"
		ptt list
	`, bashPath, tmpBin, tmpDir)

	cmd := exec.Command(bashPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// list output should contain worktree names
	if !strings.Contains(string(output), "main") {
		t.Errorf("Expected 'ptt list' output to contain 'main', got: %s", output)
	}
}

// TestZshWrapperCd tests that the zsh wrapper successfully changes directory with 'ptt cd'.
func TestZshWrapperCd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	zshPath, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not available")
	}

	tmpBin := buildPttBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		export SHELL="%s"
		eval "$('%s' shell-init)"
		cd "%s/main"
		ptt cd feature
		echo "RESULT_PWD=$PWD"
	`, zshPath, tmpBin, tmpDir)

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

// TestZshWrapperCdHome tests that the zsh wrapper successfully navigates home with 'ptt cd' (no args).
func TestZshWrapperCdHome(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	zshPath, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not available")
	}

	tmpBin := buildPttBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		export SHELL="%s"
		eval "$('%s' shell-init)"
		cd "%s/feature"
		ptt cd
		echo "RESULT_PWD=$PWD"
	`, zshPath, tmpBin, tmpDir)

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

// TestZshWrapperNew tests that the zsh wrapper successfully changes directory with 'ptt new'.
func TestZshWrapperNew(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	zshPath, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not available")
	}

	tmpBin := buildPttBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		export SHELL="%s"
		eval "$('%s' shell-init)"
		cd "%s/main"
		ptt new test-branch
		echo "RESULT_PWD=$PWD"
	`, zshPath, tmpBin, tmpDir)

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

// TestZshWrapperPassthrough tests that non-cd commands pass through correctly in zsh.
func TestZshWrapperPassthrough(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	zshPath, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not available")
	}

	tmpBin := buildPttBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		export SHELL="%s"
		eval "$('%s' shell-init)"
		cd "%s/main"
		ptt list
	`, zshPath, tmpBin, tmpDir)

	cmd := exec.Command(zshPath, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "main") {
		t.Errorf("Expected 'wt list' output to contain 'main', got: %s", output)
	}
}

// TestFishWrapperCd tests that the fish wrapper successfully changes directory with 'ptt cd'.
func TestFishWrapperCd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	fishPath, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish not available")
	}

	tmpBin := buildPttBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		set -x SHELL "%s"
		eval (%s shell-init)
		cd "%s/main"
		ptt cd feature
		echo "RESULT_PWD=$PWD"
	`, fishPath, tmpBin, tmpDir)

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

// TestFishWrapperCdHome tests that the fish wrapper successfully navigates home with 'ptt cd' (no args).
func TestFishWrapperCdHome(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	fishPath, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish not available")
	}

	tmpBin := buildPttBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		set -x SHELL "%s"
		eval (%s shell-init)
		cd "%s/feature"
		ptt cd
		echo "RESULT_PWD=$PWD"
	`, fishPath, tmpBin, tmpDir)

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

// TestFishWrapperNew tests that the fish wrapper successfully changes directory with 'ptt new'.
func TestFishWrapperNew(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell e2e test in short mode")
	}

	fishPath, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish not available")
	}

	tmpBin := buildPttBinary(t)
	tmpDir := t.TempDir()
	runSetup(t, tmpDir)

	script := fmt.Sprintf(`
		set -x SHELL "%s"
		eval (%s shell-init)
		cd "%s/main"
		ptt new test-branch
		echo "RESULT_PWD=$PWD"
	`, fishPath, tmpBin, tmpDir)

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
