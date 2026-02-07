package git

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ResolveWorktree finds a worktree by name using suffix matching
// It matches if the worktree basename ends with "-<name>" or equals "<name>" exactly
func ResolveWorktree(name string) (*Worktree, error) {
	worktrees, err := ListWorktrees()
	if err != nil {
		return nil, err
	}

	var matches []Worktree
	for _, wt := range worktrees {
		basename := filepath.Base(wt.Path)

		// Exact match or suffix match (repo-staging matches "staging")
		if basename == name || strings.HasSuffix(basename, "-"+name) {
			matches = append(matches, wt)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("worktree '%s' not found", name)
	}

	if len(matches) > 1 {
		return nil, fmt.Errorf("worktree '%s' is ambiguous", name)
	}

	return &matches[0], nil
}
