---
phase: 05-directory-commands
plan: 03
subsystem: core
tags: [git-worktree, stash, cobra, eject, rollback]

# Dependency graph
requires:
  - phase: 05-01
    provides: Shared git helpers (CurrentBranch, GetHomePath, WorktreePath, CurrentWorktreeRoot, IsInsideGitRepo)
  - phase: 04-02
    provides: Config executor with rollback support
provides:
  - wt eject command with full stash handling and rollback
  - Fallback branch detection (main/master for home, directory-suffix for non-home)
  - Stash conflict warning (non-fatal)
  - Same config flag support as wt new
affects: [shell-wrapper, directory-commands]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Stash push/pop pattern with conflict detection"
    - "Multi-step rollback (unstash, checkout back, remove worktree)"
    - "Home worktree detection via path equality check"

key-files:
  created:
    - cmd/eject.go
    - cmd/eject_test.go
  modified: []

key-decisions:
  - "Stash conflict detection: non-zero exit from stash pop triggers warning but doesn't fail command"
  - "Fallback branch logic: home worktree uses main/master, non-home uses directory suffix"
  - "Rollback strategy: each step has its own rollback path (pop stash, checkout back, remove worktree)"
  - "Config rollback: on config failure, undo branch switch and stash in addition to worktree removal"

patterns-established:
  - "Home worktree identification: srcRoot == homePath (works for both bare and regular repos)"
  - "Bare repo naming: worktrees created as {bareRoot}/{name}, first worktree follows same pattern"
  - "Stash counting: compare stash list count before/after to detect if stash occurred"

# Metrics
duration: 4min
completed: 2026-02-07
---

# Phase 05 Plan 03: wt eject Command Summary

**Eject command faithfully ports v1.0 flow with stash handling, fallback branch detection, config application, and multi-step rollback safety**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-07T22:09:16Z
- **Completed:** 2026-02-07T22:13:55Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Implemented wt eject command that stashes changes, switches to fallback branch, creates new worktree, and restores stash
- Fallback branch detection works for both home worktrees (main/master) and non-home worktrees (directory suffix)
- Stash conflict warning is non-fatal with clear user guidance
- Full rollback support at each step (unstash, checkout back, remove worktree)
- Config application uses same flags as wt new (--config, --skip-config, --copy, --symlink)

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement wt eject command** - `a7e395b` (feat)
2. **Task 2: Test wt eject flow** - `2be723c` (test)

**Plan metadata:** (pending - to be created after SUMMARY)

## Files Created/Modified
- `cmd/eject.go` - wt eject command with stash handling, fallback detection, config application, and rollback
- `cmd/eject_test.go` - Integration tests for eject flow (home worktree, custom name, detached HEAD, already-on-fallback)

## Decisions Made

**Stash conflict handling:** When `git stash pop` returns non-zero (merge conflicts), print warning to stderr but don't fail the command. The stash was still popped, just with conflicts. User can resolve before committing.

**Fallback branch detection:**
- Home worktree (srcRoot == homePath): try main, then master. Error if neither exists.
- Non-home worktree: extract directory suffix by removing repo name prefix, use as fallback branch. Error if branch doesn't exist.

**Config rollback enhancement:** When config execution fails, the executor removes the worktree directory. But we also need to undo the branch switch and stash in the source worktree, so eject adds `git checkout {originalBranch}` and `git stash pop` to the rollback path.

**Test naming convention:** Tests use bare repo with worktrees named `{bareRoot}/{branchname}` (e.g., `/tmp/test-bare/main`) to match wt's actual naming convention. This ensures the fallback branch detection logic works correctly.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

**Test setup issue (resolved):** Initial test setup created worktree named "main-wt" instead of "main". This didn't match wt's naming convention (bareRoot/branchname) and caused fallback branch detection to fail. Fixed by renaming worktree to "main" in test setup.

**Root cause:** In bare repo, worktrees are nested under bare root, not siblings with prefix. The first worktree should be named after the branch, not "main-wt".

**Resolution:** Updated setupBareRepoWithCommit to use `filepath.Join(bareRoot, "main")` instead of `filepath.Join(bareRoot, "main-wt")`. All tests now pass.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- wt eject complete, ready for wt new command (Plan 05-02)
- All directory-changing commands (goto, home, merge, rebase, eject) ready for shell wrapper integration
- Config flag support consistent across new and eject commands
- --output-path protocol established for shell coordination

---
*Phase: 05-directory-commands*
*Completed: 2026-02-07*

## Self-Check: PASSED
