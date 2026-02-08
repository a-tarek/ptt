---
phase: 10-uat-gap-closure
plan: 01
subsystem: shell-integration
tags: [shell-wrapper, stdout-stderr, cli-flags, shell-coordination]

# Dependency graph
requires:
  - phase: 06-shell-integration
    provides: Shell wrappers with --output-path protocol
  - phase: 04-configuration-system
    provides: BuildActionsFromFlags config plumbing
provides:
  - Clean stdout/stderr separation for shell wrapper coordination
  - --run inline flag for ad-hoc command execution
  - All setup progress messages on stderr only
affects: [uat, shell-integration, documentation]

# Tech tracking
tech-stack:
  added: []
  patterns: [stderr-for-progress, stdout-for-coordination]

key-files:
  created: []
  modified:
    - internal/setup/executor.go
    - internal/setup/run.go
    - cmd/new.go
    - cmd/eject.go
    - internal/setup/executor_test.go

key-decisions:
  - "Status messages go to stderr: Progress messages (Copied, Symlinked, Running) use fmt.Fprintf(os.Stderr) to prevent shell wrapper stdout capture interference"
  - "Run command output to stderr: cmd.Stdout = os.Stderr ensures run-action output doesn't leak into cd path"
  - "--run flag for convenience: Enables wt new feature --run 'npm install' without .wtconfig file for AI agents and quick workflows"

patterns-established:
  - "Shell coordination pattern: Only cd path on stdout, everything else (confirmations, progress, errors) on stderr"
  - "Test stderr capture: Tests that verify user-facing messages must capture os.Stderr, not os.Stdout"

# Metrics
duration: 4min
completed: 2026-02-08
---

# Phase 10 Plan 01: UAT Gap Closure Summary

**Fixed stdout leak causing garbled cd paths when wt new runs setup actions; added --run flag for inline command execution**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-08T09:05:19Z
- **Completed:** 2026-02-08T09:09:12Z
- **Tasks:** 3
- **Files modified:** 5

## Accomplishments

- Setup executor progress messages redirected from stdout to stderr (prevents shell wrapper cd path corruption)
- Run command output redirected to stderr (cmd.Stdout = os.Stderr)
- Added --run inline flag to wt new and wt eject commands for ad-hoc command execution
- All existing tests pass including shell E2E tests with bash/zsh wrappers

## Task Commits

Each task was committed atomically:

1. **Task 1: Redirect setup executor output from stdout to stderr** - `af66e20` (fix)
2. **Task 2: Add --run inline flag to wt new and wt eject** - `9b7c226` (feat)
3. **Task 3: Validate shell wrapper end-to-end tests pass** - `178516a` (test)

## Files Created/Modified

- `internal/setup/executor.go` - Changed fmt.Printf to fmt.Fprintf(os.Stderr) for Copied/Symlinked/Running messages
- `internal/setup/run.go` - Redirected cmd.Stdout to os.Stderr to prevent run-action output from being captured as cd path
- `cmd/new.go` - Added runFlags variable and --run flag wired through BuildActionsFromFlags
- `cmd/eject.go` - Added ejectRunFlags variable and --run flag wired through BuildActionsFromFlags
- `internal/setup/executor_test.go` - Fixed TestExecuteActionsStatusMessages to capture stderr instead of stdout

## Decisions Made

**Status messages to stderr only:**
- All setup progress messages (Copied, Symlinked, Running) now use `fmt.Fprintf(os.Stderr, ...)` instead of `fmt.Printf`
- Run-action command output goes to stderr via `cmd.Stdout = os.Stderr`
- WHY: Shell wrapper does `result=$(__WT_BIN__ --output-path "$@")` which captures ALL stdout as the cd target path. Only the cd path should be on stdout; everything else MUST go to stderr to prevent garbled paths like "Symlinked wt-bin\n/actual/path"

**--run flag for convenience:**
- Exposed run-action support via --run CLI flag on both wt new and wt eject
- WHY: AI agents and users benefit from `wt new feature --run "npm install"` without needing a .wtconfig file. The config system already supported run actions internally; this just makes it accessible inline.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed test capturing stdout instead of stderr**
- **Found during:** Task 3 (Validate shell wrapper tests)
- **Issue:** TestExecuteActionsStatusMessages captured stdout but status messages were moved to stderr in Task 1, causing test failure
- **Fix:** Changed test to capture os.Stderr instead of os.Stdout (lines 117-125 in executor_test.go)
- **Files modified:** internal/setup/executor_test.go
- **Verification:** All setup package tests pass
- **Committed in:** 178516a (separate test commit)

---

**Total deviations:** 1 auto-fixed (1 test bug)
**Impact on plan:** Test bug fix necessary for test suite to pass after stdout→stderr refactor. No scope creep.

## Issues Encountered

None. Plan executed smoothly with WIP shell wrapper changes (binary path resolution, subcommand routing) already in working tree.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- All 3 UAT gaps related to shell wrapper coordination are resolved:
  1. ✅ Binary path resolution (already in working tree via __WT_BIN__ placeholder)
  2. ✅ Stdout leak fixed (this plan - messages to stderr)
  3. ✅ Subcommand routing (already in working tree via case-based routing)
- Shell wrapper uses resolved absolute binary path via __WT_BIN__ placeholder
- Only cd-requiring commands (goto, home, new, eject) use --output-path; others pass through
- Setup progress messages go to stderr exclusively
- Ready for full UAT validation

## Self-Check: PASSED

All modified files and commit hashes verified to exist.

---
*Phase: 10-uat-gap-closure*
*Completed: 2026-02-08*
