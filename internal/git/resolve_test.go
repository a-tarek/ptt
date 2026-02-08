package git

import (
	"testing"
)

func TestFindClosestMatch(t *testing.T) {
	worktrees := []Worktree{
		{Path: "/repos/wt"},
		{Path: "/repos/wt-feat-1"},
		{Path: "/repos/wt-feat-2"},
		{Path: "/repos/wt-bugfix"},
		{Path: "/repos/wt-feature"},
	}

	tests := []struct {
		name     string
		input    string
		expected string // empty = no match expected
	}{
		// UAT failure case: "feat" should match wt-feat-1 or wt-feat-2, NOT "wt"
		{
			name:     "segment match prefers feat segment",
			input:    "feat",
			expected: "wt-feat-1", // Could also be wt-feat-2 (both score 100)
		},

		// Substring match
		{
			name:     "substring match",
			input:    "bugf",
			expected: "wt-bugfix",
		},

		// Exact segment
		{
			name:     "exact segment bugfix",
			input:    "bugfix",
			expected: "wt-bugfix",
		},

		// Levenshtein typo
		{
			name:     "typo featur -> feature",
			input:    "featur",
			expected: "wt-feature",
		},

		// No match at all
		{
			name:     "no match garbage",
			input:    "zzzzzzz",
			expected: "",
		},

		// Segment prefix
		{
			name:     "segment prefix fea",
			input:    "fea",
			expected: "wt-feat-1", // Could also be wt-feat-2 or wt-feature
		},

		// Case insensitivity
		{
			name:     "case insensitive BUGFIX",
			input:    "BUGFIX",
			expected: "wt-bugfix",
		},

		// Exact match on entire basename
		{
			name:     "exact match wt",
			input:    "wt",
			expected: "wt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindClosestMatch(tt.input, worktrees)

			if tt.expected == "" {
				if got != "" {
					t.Errorf("FindClosestMatch(%q) = %q, expected no match (empty string)", tt.input, got)
				}
				return
			}

			// For cases where multiple worktrees have the same score (e.g., wt-feat-1 and wt-feat-2),
			// we accept any of them
			if tt.input == "feat" || tt.input == "fea" {
				// Accept wt-feat-1, wt-feat-2, or wt-feature
				validMatches := []string{"wt-feat-1", "wt-feat-2", "wt-feature"}
				matched := false
				for _, valid := range validMatches {
					if got == valid {
						matched = true
						break
					}
				}
				if !matched {
					t.Errorf("FindClosestMatch(%q) = %q, expected one of %v", tt.input, got, validMatches)
				}
				return
			}

			if got != tt.expected {
				t.Errorf("FindClosestMatch(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFindClosestMatchCriticalCase(t *testing.T) {
	// This is the specific UAT failure case that MUST pass
	worktrees := []Worktree{
		{Path: "/repos/wt"},
		{Path: "/repos/wt-feat-1"},
		{Path: "/repos/wt-feat-2"},
	}

	got := FindClosestMatch("feat", worktrees)

	// CRITICAL: Must NOT return "wt"
	if got == "wt" {
		t.Fatalf("FindClosestMatch(\"feat\") returned \"wt\" - this is the bug we're fixing!")
	}

	// Must return one of the feat worktrees
	if got != "wt-feat-1" && got != "wt-feat-2" {
		t.Errorf("FindClosestMatch(\"feat\") = %q, expected \"wt-feat-1\" or \"wt-feat-2\"", got)
	}
}

func TestMatchScore(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		candidate  string
		wantScore  int
		scoreRange string // For approximate ranges
	}{
		// Exact segment match
		{
			name:      "exact segment match",
			input:     "feat",
			candidate: "wt-feat-1",
			wantScore: 100,
		},

		// Pure segment prefix (substring priority but good to document)
		{
			name:      "segment prefix also substring",
			input:     "fix",
			candidate: "wt-bugfix-123",
			wantScore: 80, // "fix" is substring, which scores higher than segment prefix
		},

		// Substring match
		{
			name:      "substring match",
			input:     "feat",
			candidate: "wt-feature",
			wantScore: 80,
		},

		// Segment prefix (also matches substring since "fea" is in "wt-feat-1")
		{
			name:      "segment prefix",
			input:     "fea",
			candidate: "wt-feat-1",
			wantScore: 80, // Substring match takes priority
		},

		// Candidate prefix (also matches substring since "wt-f" is in "wt-feat")
		{
			name:      "candidate prefix",
			input:     "wt-f",
			candidate: "wt-feat",
			wantScore: 80, // Substring match takes priority
		},

		// Levenshtein close match (also substring match since "featur" is in "feature")
		{
			name:      "levenshtein typo",
			input:     "featur",
			candidate: "feature",
			wantScore: 80, // Substring match takes priority
		},

		// No match
		{
			name:      "no match",
			input:     "xyz",
			candidate: "wt-feat",
			wantScore: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchScore(tt.input, tt.candidate)

			if tt.scoreRange != "" {
				// Score should be in expected range
				if got < 15 || got > 20 {
					t.Errorf("matchScore(%q, %q) = %d, expected range %s", tt.input, tt.candidate, got, tt.scoreRange)
				}
			} else {
				if got != tt.wantScore {
					t.Errorf("matchScore(%q, %q) = %d, expected %d", tt.input, tt.candidate, got, tt.wantScore)
				}
			}
		})
	}
}
