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

// matchScore computes a score for how well input matches candidate
// Higher score = better match. Returns 0 if no match.
func matchScore(input, candidate string) int {
	// Priority 1: Exact segment match (split on "-")
	// "feat" matches segment in "wt-feat-1" -> score 100
	segments := strings.Split(candidate, "-")
	for _, seg := range segments {
		if seg == input {
			return 100
		}
	}

	// Priority 2: Substring match
	// "feat" is substring of "wt-feature" -> score 80
	if strings.Contains(candidate, input) {
		return 80
	}

	// Priority 3: Segment prefix match
	// "fea" matches start of segment "feat" in "wt-feat-1" -> score 60
	for _, seg := range segments {
		if strings.HasPrefix(seg, input) {
			return 60
		}
	}

	// Priority 4: Input is a prefix of candidate
	// "wt-f" matches start of "wt-feat-1" -> score 40
	if strings.HasPrefix(candidate, input) {
		return 40
	}

	// Priority 5: Levenshtein distance (for typos)
	// "featur" vs "feature" -> distance 1 -> score 20
	dist := levenshteinDistance(input, candidate)
	maxDist := max(3, len(input)/2) // scale threshold with input length
	if dist <= maxDist {
		return 20 - dist // closer = higher score (max 20, min 17 for dist=3)
	}

	// Also check Levenshtein against individual segments
	// "fea" vs "feat" -> distance 1 -> score 15
	for _, seg := range segments {
		dist := levenshteinDistance(input, seg)
		if dist <= 2 {
			return 15 - dist
		}
	}

	return 0 // no match
}

// FindClosestMatch returns the closest worktree name using segment-aware scoring
// Returns empty string if no close match found (score = 0)
func FindClosestMatch(input string, worktrees []Worktree) string {
	input = strings.ToLower(input)
	var bestMatch string
	bestScore := 0 // higher is better

	for _, wt := range worktrees {
		basename := filepath.Base(wt.Path)
		basenameOriginal := basename
		basename = strings.ToLower(basename)
		score := matchScore(input, basename)
		if score > bestScore {
			bestScore = score
			bestMatch = basenameOriginal // preserve original case
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
		suggestion := FindClosestMatch(name, worktrees)
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
