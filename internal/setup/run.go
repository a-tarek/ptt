package setup

import (
	"fmt"
	"os"
	"os/exec"
)

// RunCommand executes a command via sh -c with the given working directory
// Stdout and stderr are streamed to the user in real-time
func RunCommand(workDir, command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command failed with exit code %d", exitErr.ExitCode())
		}
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}
