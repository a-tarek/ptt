package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// ListBranches returns deduplicated branch names (local + remote, prefix stripped).
func ListBranches() ([]string, error) {
	localCmd := exec.Command("git", "branch", "--format=%(refname:short)")
	localOut, err := localCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git branch failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute git: %w", err)
	}

	remoteCmd := exec.Command("git", "branch", "-r", "--format=%(refname:short)")
	remoteOut, err := remoteCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git branch -r failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute git: %w", err)
	}

	return parseBranches(string(localOut), string(remoteOut)), nil
}

// parseBranches deduplicates local and remote branch names.
// Remote branches have their remote prefix stripped (e.g. "origin/feat" -> "feat").
// HEAD pointers (e.g. "origin/HEAD") are skipped.
func parseBranches(localOutput, remoteOutput string) []string {
	seen := make(map[string]bool)
	var branches []string

	// Local branches first (take precedence)
	for _, line := range strings.Split(localOutput, "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		if !seen[name] {
			seen[name] = true
			branches = append(branches, name)
		}
	}

	// Remote branches with prefix stripped
	for _, line := range strings.Split(remoteOutput, "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}

		// Skip HEAD pointers like "origin/HEAD"
		if strings.HasSuffix(name, "/HEAD") {
			continue
		}

		// Strip remote prefix: "origin/feat/bar" -> "feat/bar"
		if idx := strings.Index(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}

		if name != "" && !seen[name] {
			seen[name] = true
			branches = append(branches, name)
		}
	}

	return branches
}
