---
phase: 15-bare-repo-infrastructure
plan: 01
subsystem: git
tags: [bare-repo, worktree, git, infrastructure, path-resolution]

# Dependency graph
requires:
  - phase: 13-shell-wrapper-rebrand
    provides: Core ptt CLI with mk/go/rm/ls commands and worktree management
provides:
  - BareRepoRoot() function for detecting ptt bare repo structure
  - ConfigRoot() function for config location resolution
  - Refactored WorktreePath() using BareRepoRoot() with backward compatibility
affects: [16-ptt-config-bare-support, 17-mk-command-bare-mode, 18-cd-command-rename, 19-mk-bare-repo-command]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Ptt bare repo convention: .bare directory + .git file at container root"
    - "BareRepoRoot() as single source of truth for bare repo detection"
    - "Backward compatibility: support both ptt bare repos and standard bare repos"

key-files:
  created: []
  modified:
    - internal/git/repo.go
    - internal/git/repo_test.go

key-decisions:
  - "BareRepoRoot() only detects ptt bare repos (.bare convention), rejects standard bare repos"
  - "WorktreePath() maintains backward compatibility with standard bare repos via fallback logic"
  - "ConfigRoot() tries BareRepoRoot() first, falls back to GetHomePath() for non-bare repos"
  - "Integration tests use real git commands to create bare repo structures (not mocks)"

patterns-established:
  - "TDD approach: RED (failing tests) → GREEN (implementation) → REFACTOR (cleanup)"
  - "Path canonicalization using filepath.EvalSymlinks for macOS /var vs /private/var"
  - "Git command detection via git rev-parse --git-common-dir for shared git data"

# Metrics
duration: 7min
completed: 2026-02-09
---

# Phase 15 Plan 01: Bare Repo Infrastructure Summary

**BareRepoRoot(), ConfigRoot(), and refactored WorktreePath() with ptt bare repo detection via git rev-parse --git-common-dir**

## Performance

- **Duration:** 7 min
- **Started:** 2026-02-09T09:01:17Z
- **Completed:** 2026-02-09T09:07:54Z
- **Tasks:** 3 (TDD: RED, GREEN, REFACTOR)
- **Files modified:** 2

## Accomplishments
- BareRepoRoot() reliably detects ptt bare repo structure from any CWD (worktree, container root, subdirectory)
- ConfigRoot() returns bare repo root in ptt bare context, home path otherwise
- WorktreePath() refactored to use BareRepoRoot() while maintaining backward compatibility with existing tests

## Task Commits

TDD cycle with three atomic commits:

1. **RED: Add failing tests** - `beeacde` (test)
   - createPttBareRepo() helper creates real bare repo structure with .bare directory
   - Tests for BareRepoRoot() from different CWDs
   - Tests for ConfigRoot() in bare and non-bare contexts
   - Tests for WorktreePath() nested mode in bare repos

2. **GREEN: Implement functions** - `084203d` (feat)
   - BareRepoRoot() using git rev-parse --git-common-dir + .bare validation
   - ConfigRoot() with BareRepoRoot() → GetHomePath() fallback
   - WorktreePath() refactored with ptt bare detection + backward compatibility

3. **REFACTOR: Clean up** - `6934412` (refactor)
   - Removed obsolete TestWorktreePath_BareRepoOld skip test

**Plan metadata commit:** (deferred to final commit with SUMMARY and STATE)

## Files Created/Modified
- `internal/git/repo.go` - Added BareRepoRoot(), ConfigRoot(); refactored WorktreePath()
- `internal/git/repo_test.go` - Added comprehensive integration tests with real git bare repos

## Decisions Made

**1. BareRepoRoot() detection algorithm**
- Uses `git rev-parse --git-common-dir` to find shared .bare directory
- Validates basename is `.bare` (ptt convention)
- Checks container root has `.git` file (not directory) via os.Lstat()
- Rejects standard bare repos (they don't follow ptt convention)

**2. Backward compatibility in WorktreePath()**
- Ptt bare repos: use BareRepoRoot() result for nested paths
- Standard bare repos: fall back to legacy detection (`.git` suffix) for nested paths
- Regular repos: use sibling path logic (existing behavior)
- Ensures all existing tests pass without modification

**3. Path canonicalization for tests**
- Used filepath.EvalSymlinks() to handle macOS /var vs /private/var symlinks
- Ensures tests pass consistently across environments

**4. Integration test approach**
- Create real bare repo structures using git commands (not mocks)
- Test helper createPttBareRepo() mirrors actual ptt bare repo creation
- Exercises full git command execution path

## Deviations from Plan

**Auto-fixed Issues:**

**1. [Rule 3 - Blocking] Added backward compatibility to WorktreePath()**
- **Found during:** GREEN phase - existing tests failed
- **Issue:** Existing tests use standard bare repos (`git init --bare repo.git`), not ptt bare repos
- **Fix:** Added fallback detection logic in WorktreePath() to support both ptt and standard bare repos
- **Files modified:** internal/git/repo.go
- **Verification:** All existing tests pass (cmd, shell integration tests)
- **Committed in:** 084203d (feat commit)

**2. [Rule 3 - Blocking] Fixed path comparison in tests for macOS**
- **Found during:** GREEN phase - tests failed on macOS
- **Issue:** macOS uses /private/var (real) and /var (symlink), direct string comparison failed
- **Fix:** Used filepath.EvalSymlinks() to canonicalize paths before comparison
- **Files modified:** internal/git/repo_test.go
- **Verification:** Tests pass on macOS
- **Committed in:** 084203d (feat commit)

---

**Total deviations:** 2 auto-fixed (both Rule 3 - Blocking)
**Impact on plan:** Both fixes were essential to maintain existing functionality while adding new ptt bare repo support. Backward compatibility ensures no breaking changes. No scope creep.

## Issues Encountered

None - TDD approach caught issues early (path canonicalization, backward compatibility) and fixed them during GREEN phase.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

✅ **Ready for Phase 16 (Ptt Config Bare Support)**
- BareRepoRoot() provides reliable detection foundation
- ConfigRoot() ready to be used for .pttconfig location resolution

✅ **Ready for Phase 17-19 (mk/cd commands, mk-bare-repo)**
- WorktreePath() supports nested mode for ptt bare repos
- Backward compatibility ensures existing workflows continue working

**No blockers or concerns.**

---
*Phase: 15-bare-repo-infrastructure*
*Completed: 2026-02-09*

## Self-Check: PASSED

All files and commits verified.
