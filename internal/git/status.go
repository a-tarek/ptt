package git

import (
	"os/exec"
	"strconv"
	"strings"
)

// StatusInfo holds parsed git status for prompt display
type StatusInfo struct {
	Branch    string
	Staged    int
	Modified  int
	Untracked int
	Ahead     int
	Behind    int
}

// Status parses `git status --porcelain -b` for the given worktree path
func Status(path string) (StatusInfo, error) {
	cmd := exec.Command("git", "-C", path, "status", "--porcelain", "-b")
	output, err := cmd.Output()
	if err != nil {
		return StatusInfo{}, err
	}

	var info StatusInfo
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}

		// Branch line: ## main...origin/main [ahead 3, behind 2]
		if strings.HasPrefix(line, "## ") {
			info.Branch = parseBranchName(line)
			info.Ahead, info.Behind = parseAheadBehind(line)
			continue
		}

		if len(line) < 2 {
			continue
		}

		x, y := line[0], line[1]

		if x == '?' && y == '?' {
			info.Untracked++
			continue
		}

		// X (index/staged): anything other than ' ' or '?' means staged
		if x != ' ' && x != '?' {
			info.Staged++
		}
		// Y (working tree): anything other than ' ' or '?' means modified
		if y != ' ' && y != '?' {
			info.Modified++
		}
	}

	return info, nil
}

// parseBranchName extracts the local branch name from the ## header line.
// Examples:
//
//	"## main...origin/main [ahead 3]" → "main"
//	"## feature/foo" → "feature/foo"
//	"## HEAD (no branch)" → ""
//	"## No commits yet on main" → "main"
func parseBranchName(line string) string {
	// Strip "## "
	rest := strings.TrimPrefix(line, "## ")

	// Detached HEAD
	if strings.HasPrefix(rest, "HEAD (no branch)") {
		return ""
	}

	// Initial commit: "No commits yet on <branch>"
	if strings.HasPrefix(rest, "No commits yet on ") {
		return strings.TrimPrefix(rest, "No commits yet on ")
	}

	// "main...origin/main [ahead 3]" or just "main"
	if idx := strings.Index(rest, "..."); idx >= 0 {
		return rest[:idx]
	}
	// No upstream: "## main" — strip trailing bracket info if any
	if idx := strings.Index(rest, " ["); idx >= 0 {
		return rest[:idx]
	}
	return rest
}

func parseAheadBehind(line string) (ahead, behind int) {
	// ## main...origin/main [ahead 3, behind 2]
	// ## main...origin/main [ahead 3]
	// ## main...origin/main [behind 2]
	// ## main...origin/main
	bracket := strings.Index(line, "[")
	if bracket < 0 {
		return 0, 0
	}
	info := line[bracket:]

	if i := strings.Index(info, "ahead "); i >= 0 {
		s := info[i+6:]
		if end := strings.IndexAny(s, ",]"); end >= 0 {
			ahead, _ = strconv.Atoi(s[:end])
		}
	}
	if i := strings.Index(info, "behind "); i >= 0 {
		s := info[i+7:]
		if end := strings.IndexAny(s, ",]"); end >= 0 {
			behind, _ = strconv.Atoi(s[:end])
		}
	}
	return ahead, behind
}
