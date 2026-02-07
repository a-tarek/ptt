---
phase: 04-configuration-system
plan: 01
subsystem: config
tags: [go, tdd, config-parser, validation, flags]

# Dependency graph
requires:
  - phase: 03-core-go-binary-foundation
    provides: Go binary foundation with git integration
provides:
  - Config parsing library (action types, parser, validator)
  - Config resolution (bare names, exact paths, default)
  - Inline flag builder and duplicate detection
  - Init command writing to repo root with --name flag
affects: [04-02-setup-actions, 05-worktree-creation, 06-completion]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "TDD with RED-GREEN-REFACTOR cycle"
    - "Upfront validation with collected error reporting"
    - "Config resolution with bare name vs exact path detection"

key-files:
  created:
    - internal/config/action.go
    - internal/config/parser.go
    - internal/config/parser_test.go
    - internal/config/validator.go
    - internal/config/validator_test.go
    - internal/config/resolve.go
    - internal/config/resolve_test.go
    - internal/config/flags.go
    - internal/config/flags_test.go
  modified:
    - cmd/init_cmd.go

key-decisions:
  - "Parser uses SplitN(line, ' ', 2) to preserve spaces in run commands"
  - "Validator reports all errors at once (not fail-on-first)"
  - "Bare names resolve to .wtconfig-{name}, paths with '/' treated as exact paths"
  - "Inline flags ordered by type (Cobra limitation), not interleaved"
  - "Init command writes to repo root, not cwd"

patterns-established:
  - "Config Action struct with Type, Path, Line fields for error reporting"
  - "ParseFile -> ValidateActions -> Execute pattern for config processing"
  - "CheckDuplicatePaths for inline flags validates before execution"

# Metrics
duration: 5min
completed: 2026-02-07
---

# Phase 4 Plan 01: Config Parser & Validator Summary

**TDD-built config library parsing .wtconfig files with SplitN for space-preserving run commands, upfront validation collecting all errors, bare-name resolution to .wtconfig-{name}, and init command writing to repo root with --name flag support**

## Performance

- **Duration:** 5 min
- **Started:** 2026-02-07T12:46:07Z
- **Completed:** 2026-02-07T12:50:48Z
- **Tasks:** 3
- **Files created:** 9
- **Files modified:** 1

## Accomplishments
- TDD-built config parsing library with 36 comprehensive tests
- Parser handles comments, blank lines, whitespace, run commands with spaces using SplitN
- Validator reports all errors at once (not fail-on-first)
- Config resolution maps bare names to .wtconfig-{name}, treats paths with "/" as exact
- Inline flag builder creates ordered actions with duplicate detection
- Init command writes to repo root with --name flag for named variants

## Task Commits

Each task was committed atomically:

1. **Task 1: Config types and parser with tests (TDD)** - `b2eea94` (test/feat - combined TDD RED+GREEN)
2. **Task 2: Validation, resolution, and inline flags with tests (TDD)** - `fab5b20` (test/feat - combined TDD RED+GREEN)
3. **Task 3: Update init command for repo root and --name flag** - `0cda6cf` (feat)

_Note: TDD tasks combined RED+GREEN phases into single commits for atomic delivery_

## Files Created/Modified

**Created:**
- `internal/config/action.go` - Action struct and constants (copy, symlink, run)
- `internal/config/parser.go` - ParseFile function reading .wtconfig into []Action
- `internal/config/parser_test.go` - 10 tests for config parsing edge cases
- `internal/config/validator.go` - ValidateActions with collected error reporting
- `internal/config/validator_test.go` - 9 tests for upfront validation
- `internal/config/resolve.go` - ResolveConfigPath for --config flag resolution
- `internal/config/resolve_test.go` - 7 tests for config path resolution
- `internal/config/flags.go` - BuildActionsFromFlags and CheckDuplicatePaths
- `internal/config/flags_test.go` - 10 tests for inline flag handling

**Modified:**
- `cmd/init_cmd.go` - Updated to write to repo root with --name flag support

## Decisions Made

**Parser implementation:**
- Used `strings.SplitN(line, " ", 2)` instead of `strings.Split()` to preserve spaces in run commands (critical for `run npm install --save-dev typescript`)
- Line numbers tracked in Action struct for error reporting
- Comments and blank lines skipped with `strings.TrimSpace()` + prefix check

**Validator design:**
- Collect all errors in `[]string` before returning (not fail-on-first)
- Check copy/symlink source files exist via `os.Stat()`
- Accept run commands without file existence checks (command strings, not paths)
- Reject empty run commands and unknown action types

**Config resolution:**
- Bare name (no "/") → `filepath.Join(repoRoot, ".wtconfig-"+name)`
- Contains "/" → treated as exact path (relative or absolute)
- Empty name → `filepath.Join(repoRoot, ".wtconfig")`
- Return error for nonexistent configs (no silent fallback)

**Inline flags:**
- Order: all copies, then all symlinks, then all runs (Cobra limitation - cannot preserve interleaved order across flag types)
- Line numbers set to 0 (not from file)
- CheckDuplicatePaths validates before execution to catch errors early

**Init command:**
- Write to `git.CurrentWorktreeRoot()` not `os.Getwd()` (matches v1.0 and locked decision)
- --name flag creates `.wtconfig-{name}` variant
- Error messages include actual filename (`.wtconfig` vs `.wtconfig-ci`)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

**Test resolution test initially failed:**
- TestResolveConfigPath_ExactPath created file at absolute path but expected relative path return
- Fixed by correcting test expectations to match implementation: exact paths are returned as-is
- Verification: All 36 tests pass

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

**Ready for Phase 4 Plan 02 (Setup Actions):**
- Config parsing library complete and tested
- Parser handles all three action types (copy, symlink, run)
- Validator provides upfront error checking
- Resolver handles bare names and exact paths
- Flag builder creates actions from CLI flags
- Init command writes configs to repo root

**What Plan 02 will consume:**
- `config.ParseFile(path)` to read .wtconfig files
- `config.ValidateActions(srcRoot, actions)` to check before execution
- `config.ResolveConfigPath(repoRoot, name)` for --config flag
- `config.BuildActionsFromFlags(...)` for inline flags
- `config.CheckDuplicatePaths(...)` for duplicate detection

**No blockers or concerns.**

---
*Phase: 04-configuration-system*
*Completed: 2026-02-07*

## Self-Check: PASSED

All created files verified to exist.
All commit hashes verified in git log.
