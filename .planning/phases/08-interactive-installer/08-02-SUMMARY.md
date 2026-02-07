---
phase: 08-interactive-installer
plan: 02
subsystem: cli
tags: [cobra, installer, shell-integration, uninstall]

# Dependency graph
requires:
  - phase: 08-01
    provides: installer package with marker block operations (HasMarkerBlock, RemoveMarkerBlock, BackupFile, RestoreBackup)
provides:
  - wt uninstall command with confirmation flow and rc file cleanup
  - Clean removal path for wt shell integration
  - Paired install/uninstall commands for complete lifecycle
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Confirmation prompts before destructive operations
    - Backup before modify pattern

key-files:
  created:
    - cmd/uninstall.go
  modified: []

key-decisions:
  - "No automatic uncommenting of v1 lines - user can manually uncomment if reverting to v1"
  - "Uninstall only cleans rc file, does not run npm uninstall (would be self-destructive mid-execution)"
  - "Print npm uninstall instructions for user to complete uninstallation"

patterns-established:
  - "Paired install/uninstall commands for shell integration lifecycle"
  - "Display ~-prefixed paths for better UX"

# Metrics
duration: 2min
completed: 2026-02-07
---

# Phase 08 Plan 02: Uninstall Command Summary

**wt uninstall command with confirmation flow: detects shell, shows marker block preview, backs up rc file, removes integration cleanly, prints npm uninstall instructions**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-07T21:48:23Z
- **Completed:** 2026-02-07T21:49:15Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments
- Created `wt uninstall` command that pairs with `wt install` for complete lifecycle
- Confirmation flow shows exactly what will be removed before proceeding
- Backup/rollback mechanism protects user's rc file
- Clean handling of not-installed case (no-op exit)
- Clear npm uninstall instructions after rc file cleanup

## Task Commits

Each task was committed atomically:

1. **Task 1: Create wt uninstall command with confirmation and cleanup** - `5ccc3b6` (feat)

## Files Created/Modified
- `cmd/uninstall.go` - wt uninstall cobra command with confirmation flow, marker block removal, backup/rollback, and npm uninstall instructions

## Decisions Made

**No automatic uncommenting of v1 lines:**
- When user runs uninstall, we do NOT uncomment v1 lines that were commented during install
- Rationale: Automatic uncommenting could be fragile, user can manually uncomment if they want to revert to v1
- This is a deliberate choice for safety and simplicity

**Uninstall only cleans rc file:**
- The command does NOT run `npm uninstall -g @potato/wt`
- Rationale: User is running `wt uninstall` which means the binary is still installed; removing the npm package mid-execution would be self-destructive
- Instead, print clear instructions for the user to complete uninstallation

**Print npm uninstall instructions:**
- After successfully removing marker block, print: "To complete uninstallation: 1. Restart terminal, 2. Run npm uninstall command"
- Rationale: Makes it clear that rc file cleanup is not complete uninstallation, guides user to finish

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Phase 8 (Interactive Installer) is now complete:
- `wt install` (08-01) - guided setup with v1 migration
- `wt uninstall` (08-02) - clean removal path

Both commands reuse the installer package (internal/installer/rcfile.go, internal/installer/paths.go) for marker block operations, shell detection, and rc file management.

Ready for Phase 9 (Remaining Features) - any additional polish or remaining v1 parity features.

---
*Phase: 08-interactive-installer*
*Completed: 2026-02-07*

## Self-Check: PASSED
