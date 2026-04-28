package staleness

import (
	"testing"
	"time"

	"github.com/a-tarek/ptt/internal/git"
)

func TestParseDuration(t *testing.T) {
	hour := time.Hour
	day := 24 * hour
	week := 7 * day
	month := 30 * day

	tests := []struct {
		in      string
		want    time.Duration
		wantErr bool
	}{
		{"30d", 30 * day, false},
		{"30", 30 * day, false}, // bare number defaults to days
		{"2w", 2 * week, false},
		{"3m", 3 * month, false},
		{"12h", 12 * hour, false},
		{"5min", 5 * time.Minute, false},
		{"45s", 45 * time.Second, false},
		{"1.5d", 36 * hour, false},
		{"", 0, true},
		{"abc", 0, true},
		{"d", 0, true},  // missing number
		{"5x", 0, true}, // unknown unit
	}
	for _, tt := range tests {
		got, err := ParseDuration(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseDuration(%q): expected error, got %v", tt.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseDuration(%q): unexpected error %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseDuration(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestCriteria_Empty(t *testing.T) {
	if !(Criteria{}).Empty() {
		t.Error("zero-value Criteria should be Empty")
	}
	if (Criteria{MergedBase: "main"}).Empty() {
		t.Error("Criteria with MergedBase should not be Empty")
	}
	if (Criteria{Stale: time.Hour}).Empty() {
		t.Error("Criteria with Stale should not be Empty")
	}
	if (Criteria{Gone: true}).Empty() {
		t.Error("Criteria with Gone should not be Empty")
	}
	if (Criteria{NoUpstream: true}).Empty() {
		t.Error("Criteria with NoUpstream should not be Empty")
	}
}

func TestEvaluate_GoneFilter(t *testing.T) {
	now := time.Now()
	wt := git.Worktree{Path: "/tmp/nonexistent-worktree", Branch: "feat-x"}

	// Gone filter when upstream lookup is unknown should not match.
	c := Criteria{Gone: true}
	d := Evaluate(wt, c, now)
	if d.Match {
		t.Error("expected no match for unknown upstream with --gone")
	}
}

func TestEvaluate_StaleFilter(t *testing.T) {
	now := time.Now()
	wt := git.Worktree{Path: "/tmp/nonexistent-worktree", Branch: "feat-x"}

	// HeadModTime fails (path doesn't exist) → LastEdited zero → no match for stale.
	c := Criteria{Stale: 30 * 24 * time.Hour}
	d := Evaluate(wt, c, now)
	if d.Match {
		t.Error("expected no match when LastEdited is zero")
	}
}

func TestEvaluate_AndSemantics(t *testing.T) {
	now := time.Now()
	wt := git.Worktree{Path: "/tmp/nonexistent-worktree", Branch: "feat-x"}

	// Two filters: both must match. With both unresolvable, neither matches → overall no match.
	c := Criteria{Gone: true, Stale: 30 * 24 * time.Hour}
	d := Evaluate(wt, c, now)
	if d.Match {
		t.Error("AND semantics should fail when any single criterion fails")
	}
	if len(d.Reasons) != 0 {
		t.Errorf("expected no reasons on no-match, got %v", d.Reasons)
	}
}
