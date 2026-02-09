---
phase: 19-polish
plan: 01
subsystem: ui
tags: [worktree, list, bare-repo, filter]

# Dependency graph
requires:
  - phase: 15-bare-repo-infra
    provides: IsBare field on Worktree struct from git.ListWorktrees()
provides:
  - Bare repository metadata entries filtered from ptt ls output
  - Users only see real worktrees, not internal .bare directory
affects: [polish, user-experience]

# Tech tracking
tech-stack:
  added: []
  patterns: [Filter bare entries in presentation layer, not data layer]

key-files:
  created: [cmd/list_test.go]
  modified: [cmd/list.go]

key-decisions:
  - "Filter at presentation layer - cmd/list.go filters IsBare entries after fetching, not in internal/git/worktree.go"
  - "Use IsBare boolean flag - no path-based string matching for reliability"
  - "Re-check after filtering - handle edge case of bare-only repo producing empty list"

patterns-established:
  - "Presentation-layer filtering: Data retrieval functions return complete data, display functions filter for UX"

# Metrics
duration: 4min
completed: 2026-02-09
---

# Phase 19 Plan 01: Filter Bare Entries from ls Output Summary

**ptt ls now filters .bare metadata directory from output, showing only real worktrees to users**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-09T14:11:46Z
- **Completed:** 2026-02-09T14:15:49Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Users running `ptt ls` in bare repos no longer see confusing `.bare` metadata entry
- Filter uses IsBare boolean flag for reliable detection across all platforms
- Integration tests verify bare entries excluded while normal repos work unchanged

## Task Commits

Each task was committed atomically:

1. **Task 1: Add bare entry filter to ptt ls command** - `c34e4b2` (feat)
2. **Task 2: Add integration test for bare entry filtering** - `cd97158` (test)

## Files Created/Modified
- `cmd/list.go` - Added filter loop after empty check to exclude IsBare worktrees with re-check for empty list
- `cmd/list_test.go` - Integration tests with real git commands verifying .bare filtered while main worktree displayed

## Decisions Made

**1. Filter at presentation layer**
- Filtering happens in cmd/list.go after fetching worktrees, not in internal/git/worktree.go
- Rationale: internal/git should return complete data; display commands decide what to show
- Maintains separation of concerns

**2. Use IsBare boolean flag**
- Filter checks `wt.IsBare` field, not path-based string matching like `strings.Contains(wt.Path, ".bare")`
- Rationale: IsBare flag is reliable across all git configurations and path variations
- Path-based detection could break with different naming conventions

**3. Re-check after filtering**
- Added second empty check after filtering bare entries
- Rationale: Edge case where bare-only repo (no real worktrees) should produce empty output, not error

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

**Testing with stdout capture**
- Initial attempt to capture cobra command output via `cmd.SetOut(buf)` failed
- Issue: list.go uses `fmt.Printf` which writes directly to os.Stdout, not cobra's output writer
- Resolution: Redirect os.Stdout to pipe during test execution to capture output
- Note: Future improvement could use `cmd.OutOrStdout()` pattern, but that's beyond this plan's scope

## Next Phase Readiness

- ls command ready for production use in bare repos
- Filter pattern established for other commands that may need similar UX improvements
- No blockers for remaining polish phase work

---
*Phase: 19-polish*
*Completed: 2026-02-09*

## Self-Check: PASSED
