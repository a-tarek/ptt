package git

import (
	"fmt"
	"path/filepath"
	"strings"
)

// levenshteinDistance computes the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	// Handle empty strings
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			deletion := matrix[i-1][j] + 1
			insertion := matrix[i][j-1] + 1
			substitution := matrix[i-1][j-1] + cost

			min := deletion
			if insertion < min {
				min = insertion
			}
			if substitution < min {
				min = substitution
			}

			matrix[i][j] = min
		}
	}

	return matrix[len(s1)][len(s2)]
}

// findClosestMatch returns the closest worktree name based on Levenshtein distance
// Returns empty string if no close match found (distance > 3)
func findClosestMatch(input string, worktrees []Worktree) string {
	const maxDistance = 3
	var bestMatch string
	bestDistance := maxDistance + 1

	for _, wt := range worktrees {
		basename := filepath.Base(wt.Path)
		distance := levenshteinDistance(input, basename)

		if distance <= maxDistance && distance < bestDistance {
			bestDistance = distance
			bestMatch = basename
		}
	}

	return bestMatch
}

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
		// Try to find a close match using fuzzy matching
		suggestion := findClosestMatch(name, worktrees)
		if suggestion != "" {
			return nil, fmt.Errorf("worktree '%s' not found. Did you mean '%s'?", name, suggestion)
		}
		return nil, fmt.Errorf("worktree '%s' not found", name)
	}

	if len(matches) > 1 {
		// List all matching names
		names := make([]string, len(matches))
		for i, wt := range matches {
			names[i] = filepath.Base(wt.Path)
		}
		return nil, fmt.Errorf("worktree '%s' is ambiguous (matches: %s)", name, strings.Join(names, ", "))
	}

	return &matches[0], nil
}
