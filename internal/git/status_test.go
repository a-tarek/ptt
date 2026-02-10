package git

import "testing"

func TestParseBranchName(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{"with upstream", "## main...origin/main", "main"},
		{"with upstream and ahead", "## main...origin/main [ahead 3]", "main"},
		{"with upstream and behind", "## dev...origin/dev [behind 2]", "dev"},
		{"with upstream ahead and behind", "## main...origin/main [ahead 3, behind 2]", "main"},
		{"feature branch slash", "## feature/foo...origin/feature/foo", "feature/foo"},
		{"no upstream", "## main", "main"},
		{"detached HEAD", "## HEAD (no branch)", ""},
		{"initial commit", "## No commits yet on main", "main"},
		{"initial commit feature", "## No commits yet on feature/bar", "feature/bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBranchName(tt.line)
			if got != tt.want {
				t.Errorf("parseBranchName(%q) = %q, want %q", tt.line, got, tt.want)
			}
		})
	}
}

func TestParseAheadBehind(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		wantAhead  int
		wantBehind int
	}{
		{"no tracking info", "## main...origin/main", 0, 0},
		{"ahead only", "## main...origin/main [ahead 3]", 3, 0},
		{"behind only", "## main...origin/main [behind 2]", 0, 2},
		{"ahead and behind", "## main...origin/main [ahead 3, behind 2]", 3, 2},
		{"no upstream", "## main", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ahead, behind := parseAheadBehind(tt.line)
			if ahead != tt.wantAhead || behind != tt.wantBehind {
				t.Errorf("parseAheadBehind(%q) = (%d, %d), want (%d, %d)",
					tt.line, ahead, behind, tt.wantAhead, tt.wantBehind)
			}
		})
	}
}
