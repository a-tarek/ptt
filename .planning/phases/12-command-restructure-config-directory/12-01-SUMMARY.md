---
phase: 12-command-restructure-config-directory
plan: 01
subsystem: cli
tags: [cobra, aliases, command-restructure]

requires:
  - phase: 11-go-module-binary-rename
    provides: "ptt binary name and module path"
provides:
  - "mk/go/rm/ls as primary command names"
  - "backward-compatible aliases (new, goto, home, delete, list)"
  - "merged go command (goto+home behavior)"
affects: [13-shell-wrappers, 14-documentation]

tech-stack:
  added: []
  patterns:
    - "Cobra Aliases for backward-compatible command renames"
    - "MaximumNArgs for optional argument commands"

key-files:
  created: []
  modified:
    - "cmd/new.go"
    - "cmd/goto.go"
    - "cmd/delete.go"
    - "cmd/list.go"
    - "cmd/root.go"
    - "cmd/new_test.go"

key-decisions:
  - "Removed cmd/home.go entirely — merged into go command's zero-args branch"
  - "Used Cobra Aliases (not Deprecated) for full backward compatibility without warnings"

patterns-established:
  - "Cobra Aliases for command renames: primary short name + long alias"

duration: 3min
completed: 2026-02-08
---

# Phase 12 Plan 01: Command Restructure Summary

**Renamed ptt commands to Unix-style short names (mk/go/rm/ls) with Cobra aliases preserving all old names**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-08
- **Completed:** 2026-02-08
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- All four command renames: mk (new), go (goto+home), rm (delete), ls (list)
- Merged goto+home into single `go` command with MaximumNArgs(1)
- Deleted cmd/home.go — functionality absorbed into go command's zero-args branch
- Updated root help to show new command names
- All tests updated to use new primary names

## Task Commits

Each task was committed atomically:

1. **Task 1: Rename commands and merge goto+home into go** - `520823d` (feat)
2. **Task 2: Update tests for new command names** - `774e50f` (test)

## Files Created/Modified
- `cmd/new.go` - Use: "mk", Aliases: ["new"]
- `cmd/goto.go` - Use: "go", Aliases: ["goto", "home"], MaximumNArgs(1)
- `cmd/home.go` - Deleted (merged into go command)
- `cmd/delete.go` - Use: "rm", Aliases: ["delete"]
- `cmd/list.go` - Use: "ls", Aliases: ["list"]
- `cmd/root.go` - Updated Long description with new command names
- `cmd/new_test.go` - Updated SetArgs to use "mk" instead of "new"

## Decisions Made
- Removed cmd/home.go entirely rather than keeping it as a redirect — cleaner codebase, Cobra alias handles backward compat
- Used Cobra Aliases field (not Deprecated) so old names work silently without warnings

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Command restructure complete, ready for 12-02 (config directory migration)
- Shell wrappers still use old names — Phase 13 scope

---
*Phase: 12-command-restructure-config-directory*
*Completed: 2026-02-08*
