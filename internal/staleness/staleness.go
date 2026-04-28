package staleness

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/a-tarek/ptt/internal/git"
)

// Criteria holds the filter conditions for selecting worktrees.
// All non-zero criteria must match (AND semantics).
type Criteria struct {
	// MergedBase, when non-empty, requires the branch to be merged into this base.
	MergedBase string
	// Stale, when non-zero, requires the worktree HEAD's mtime to be older than this.
	Stale time.Duration
	// Gone, when true, requires the branch's upstream to be configured but deleted on remote.
	Gone bool
	// NoUpstream, when true, requires the branch to have no upstream configured at all.
	NoUpstream bool
}

// Empty reports whether no criteria are set. Callers use this to decide whether
// filtering is in effect at all.
func (c Criteria) Empty() bool {
	return c.MergedBase == "" && c.Stale == 0 && !c.Gone && !c.NoUpstream
}

// Decision is the result of evaluating a single worktree against a Criteria.
type Decision struct {
	// Match is true iff every active criterion matched.
	Match bool
	// Reasons holds short human-readable strings for each matched criterion (for display).
	Reasons []string
	// LastEdited is the HEAD mtime; zero if it could not be read.
	LastEdited time.Time
	// Upstream is the branch's upstream status.
	Upstream git.UpstreamStatus
	// Merged is true iff the branch is merged into MergedBase (only meaningful if criterion set).
	Merged bool
}

// Evaluate inspects a worktree against criteria. Errors from individual git lookups
// are tolerated (treated as "criterion did not match") so a single broken worktree
// can't poison a bulk operation.
func Evaluate(wt git.Worktree, c Criteria, now time.Time) Decision {
	d := Decision{Match: true}

	// Always read upstream status — it's cheap and several criteria use it.
	if status, err := git.BranchUpstreamStatus(wt.Branch); err == nil {
		d.Upstream = status
	}

	// Always read last-edited — useful for display even when --stale isn't active.
	if t, err := git.HeadModTime(wt.Path); err == nil {
		d.LastEdited = t
	}

	if c.MergedBase != "" {
		merged, _ := git.IsMergedInto(wt.Branch, c.MergedBase)
		d.Merged = merged
		if !merged {
			d.Match = false
		} else {
			d.Reasons = append(d.Reasons, "merged into "+c.MergedBase)
		}
	}

	if c.Stale > 0 {
		if d.LastEdited.IsZero() || now.Sub(d.LastEdited) < c.Stale {
			d.Match = false
		} else {
			d.Reasons = append(d.Reasons, "stale "+humanizeDuration(now.Sub(d.LastEdited)))
		}
	}

	if c.Gone {
		if d.Upstream != git.UpstreamGone {
			d.Match = false
		} else {
			d.Reasons = append(d.Reasons, "upstream gone")
		}
	}

	if c.NoUpstream {
		if d.Upstream != git.UpstreamNone {
			d.Match = false
		} else {
			d.Reasons = append(d.Reasons, "no upstream")
		}
	}

	return d
}

// ParseDuration extends Go's time.ParseDuration with d (24h), w (7d), and m (30d).
// Accepts the standard ns/us/ms/s/m/h units too — but here `m` means month, not minute.
// Use `min` for minutes if needed (we don't expect that resolution for worktree TTLs).
func ParseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}

	// Find the longest numeric prefix.
	i := 0
	for i < len(s) && (s[i] == '.' || (s[i] >= '0' && s[i] <= '9')) {
		i++
	}
	if i == 0 {
		return 0, fmt.Errorf("invalid duration %q: missing number", s)
	}
	num, err := strconv.ParseFloat(s[:i], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", s, err)
	}
	unit := strings.ToLower(strings.TrimSpace(s[i:]))

	switch unit {
	case "", "d":
		return time.Duration(num * 24 * float64(time.Hour)), nil
	case "w":
		return time.Duration(num * 7 * 24 * float64(time.Hour)), nil
	case "m":
		return time.Duration(num * 30 * 24 * float64(time.Hour)), nil
	case "h":
		return time.Duration(num * float64(time.Hour)), nil
	case "min":
		return time.Duration(num * float64(time.Minute)), nil
	case "s":
		return time.Duration(num * float64(time.Second)), nil
	}
	return 0, fmt.Errorf("invalid duration %q: unknown unit %q (use d, w, m, h, min, s)", s, unit)
}

// humanizeDuration renders a duration as a short human-readable string.
// Picks the largest meaningful unit so output stays compact (e.g. "5d", "2w", "3m").
func humanizeDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	day := 24 * time.Hour
	week := 7 * day
	month := 30 * day

	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < day:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < week:
		return fmt.Sprintf("%dd", int(d/day))
	case d < month:
		return fmt.Sprintf("%dw", int(d/week))
	default:
		return fmt.Sprintf("%dmo", int(d/month))
	}
}
