---
phase: 02-user-facing-documentation
plan: 02
subsystem: documentation
tags: [readme, cli, commands]

# Dependency graph
requires:
  - phase: 02-01
    provides: README structure with installation, quick start, wt init, and tab completion
provides:
  - Complete command reference for wt new (all flags, override behavior, 5 examples)
  - Complete command reference for wt eject (stash/fallback behavior, 3 examples)
  - Navigation commands documented (wt goto, wt home, wt list)
affects: [02-03]

# Tech tracking
tech-stack:
  added: []
  patterns: []

key-files:
  created: []
  modified: [README.md]

key-decisions:
  - "Document flag override precedence (flags override .wtconfig for same path)"
  - "Show one-off flag usage (paths not in .wtconfig)"
  - "Explain eject fallback branch logic (main/master vs original branch)"

patterns-established:
  - "Command reference format: usage line → description → args → flags → behavior details → examples"

# Metrics
duration: 1min
completed: 2026-02-07
---

# Phase 02 Plan 02: Command Reference Summary

**Complete command reference for wt new, wt eject, wt goto, wt home, and wt list with all flags, positional arguments, override behavior, and practical examples**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-07T05:36:28Z
- **Completed:** 2026-02-07T05:37:44Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments
- Documented wt new with --copy/--symlink flag override precedence and one-off usage (5 examples)
- Documented wt eject with complete stash/fallback/restore workflow (3 examples)
- Documented navigation commands (wt goto, wt home, wt list) with suffix matching and output format

## Task Commits

Each task was committed atomically:

1. **Task 1: Add wt new and wt eject command reference** - `094d1df` (docs)
2. **Task 2: Add wt goto, wt home, and wt list command reference** - `e3d1b67` (docs)

**Plan metadata:** (to be committed after SUMMARY.md creation)

## Files Created/Modified
- `README.md` - Added detailed command reference sections for 5 commands (wt new, wt eject, wt goto, wt home, wt list)

## Decisions Made

**1. Document flag override precedence explicitly**
- Flags take precedence over .wtconfig entries for the same path
- Flags can also specify paths NOT in .wtconfig for one-off operations
- Rationale: Users need to understand both .wtconfig defaults and CLI override capability

**2. Show wt eject fallback branch logic**
- Main/master for home worktree
- Original branch for non-home worktrees
- Rationale: Users need to understand where they'll land after ejecting

**3. Explain wt goto name resolution**
- Suffix matching (e.g., "staging" matches "myapp-staging")
- Rationale: Users need to understand how the tool resolves worktree names

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Ready for plan 02-03 (Configuration Reference). Command reference sections now include forward references to .wtconfig section (to be documented in 02-03).

## Self-Check: PASSED

---
*Phase: 02-user-facing-documentation*
*Completed: 2026-02-07*
