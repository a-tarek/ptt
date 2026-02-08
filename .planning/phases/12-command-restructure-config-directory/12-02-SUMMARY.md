---
phase: 12-command-restructure-config-directory
plan: 02
subsystem: config
tags: [config, pttconfig, directory-structure, backward-compatibility]

requires:
  - phase: 12-command-restructure-config-directory
    plan: 01
    provides: "Command restructure (mk/go/rm/ls)"
provides:
  - ".pttconfig/ directory structure for named configs"
  - "Backward-compatible .wtconfig fallback"
  - "init command creates .pttconfig/default or .pttconfig/<name>"
affects: [13-shell-wrappers, 14-documentation]

tech-stack:
  added: []
  patterns:
    - "Priority-ordered path resolution with fallback"
    - "Directory-based configuration organization"

key-files:
  created: []
  modified:
    - "internal/config/resolve.go"
    - "internal/config/resolve_test.go"
    - "cmd/init_cmd.go"
    - "cmd/new.go"
    - "cmd/eject.go"

key-decisions:
  - "Use .pttconfig/ directory for new configs, fall back to .wtconfig for legacy"
  - "init creates .pttconfig/default by default, .pttconfig/<name> with --config flag"
  - "ResolveConfigPath tries new paths first, falls back to legacy paths"

patterns-established:
  - "Priority-ordered candidate path resolution"
  - "Graceful fallback for backward compatibility"

duration: 2min
completed: 2026-02-08
---

# Phase 12 Plan 02: Config Directory Migration Summary

**Migrated configuration from flat .wtconfig files to .pttconfig/ directory with backward-compatible fallback**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-08
- **Completed:** 2026-02-08
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Updated `ResolveConfigPath` to use priority-ordered resolution:
  - Default config: `.pttconfig/default` → `.wtconfig` (fallback)
  - Named config: `.pttconfig/{name}` → `.wtconfig-{name}` (fallback)
  - Exact path (contains "/"): unchanged behavior
- Rewrote `ptt init` to create `.pttconfig/` directory structure
- Renamed flag from `--name` to `--config` for consistency across commands
- Updated all flag descriptions to reference `.pttconfig/<name>` format
- Added 6 new tests for directory-based resolution and priority
- All existing tests still pass (fallback behavior verified)

## Task Commits

Each task was committed atomically:

1. **Task 1: Update config resolution for .pttconfig/ directory with fallback** - `f9014ac` (feat)
2. **Task 2: Update init command and config flag descriptions** - `afacf09` (feat)

## Files Created/Modified

- `internal/config/resolve.go` - Priority-ordered path resolution with candidates array
- `internal/config/resolve_test.go` - Added 6 new tests for .pttconfig/ resolution
- `cmd/init_cmd.go` - Creates `.pttconfig/` directory, flag renamed to `--config`
- `cmd/new.go` - Updated --config flag description to reference `.pttconfig/<name>`
- `cmd/eject.go` - Updated --config flag description and display strings

## Decisions Made

- **Priority-ordered fallback:** Try `.pttconfig/` first, fall back to `.wtconfig` if not found
- **init creates directory structure:** `ptt init` creates `.pttconfig/default`, `ptt init --config ci` creates `.pttconfig/ci`
- **Flag consistency:** Changed init's `--name` flag to `--config` to match mk/eject
- **Display strings in eject:** Show `.pttconfig/` paths in output messages for clarity

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - migration is transparent to existing users with `.wtconfig` files.

## Next Phase Readiness

- Config directory structure complete, ready for Phase 13 (Shell Wrappers)
- Legacy `.wtconfig` files continue working during transition
- Users can migrate at their own pace by running `ptt init`

## Self-Check: PASSED

All files created/modified as documented, all commits exist in git history.

---
*Phase: 12-command-restructure-config-directory*
*Completed: 2026-02-08*
