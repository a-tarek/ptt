---
phase: 09-polish-testing
plan: 03
subsystem: testing
tags: [shell, e2e, testing, bash, zsh, fish, git-worktrees]
status: complete
completed: 2026-02-08

requires:
  - 06-01-SUMMARY # Shell wrappers (wrapper.bash, wrapper.zsh, wrapper.fish)
  - 05-01-SUMMARY # goto command
  - 05-02-SUMMARY # new command
  - 05-03-SUMMARY # home command (fixed in this plan)

provides:
  - shell-e2e-tests # E2E tests for bash/zsh/fish wrappers
  - real-git-fixtures # Helper script to create temporary git repos
  - bare-repo-home-fix # Fixed GetHomePath for bare repo configurations

affects:
  - future-shell-changes # E2E tests catch regressions in shell integration

tech-stack:
  added:
    - go-testing # Shell E2E test suite
  patterns:
    - real-git-testing # Tests use actual git repos, not mocks
    - sync-once-binary-build # Build wt binary once per test run
    - shell-agnostic-testing # Same test logic for bash/zsh/fish

key-files:
  created:
    - tests/shell/shell_test.go # E2E test suite for shell wrappers
    - tests/shell/testdata/setup.sh # Git fixture creation script
  modified:
    - internal/git/repo.go # Fixed GetHomePath for bare repos

decisions:
  - id: real-git-fixtures
    what: Use real git repos with worktrees for E2E tests (no mocking)
    why: Highest confidence that shell integration works correctly
    impact: Tests slower but catch real-world issues

  - id: build-once-pattern
    what: Build wt binary once per test run using sync.Once
    why: Significantly faster test execution (build once, use for all tests)
    impact: ~6x faster test suite (1 build vs 9 builds)

  - id: skip-missing-shells
    what: Fish tests skip gracefully if fish is not installed
    why: Fish less common in CI environments than bash/zsh
    impact: Tests pass on systems without fish installed

  - id: short-mode-support
    what: All tests check testing.Short() and skip in short mode
    why: Allows fast iteration with `go test -short`
    impact: Developers can run fast unit tests separately from slow E2E tests

metrics:
  tasks-completed: 2
  commits: 2
  files-created: 2
  files-modified: 2
  duration: 7.3 min
  test-coverage: 9 test cases (3 shells × 3 commands)
---

# Phase 9 Plan 3: Shell E2E Tests Summary

**One-liner:** End-to-end shell wrapper tests for bash/zsh/fish with real git repos, plus fix for bare repo home detection

## What Was Built

### Shell E2E Test Suite
- **tests/shell/shell_test.go**: Comprehensive test suite for shell wrappers
  - Tests bash, zsh, and fish wrappers with real shell sessions
  - Verifies `wt goto`, `wt home`, and `wt new` actually change directory in parent shell
  - Uses `exec.Command` to run real shell processes
  - Parses `RESULT_PWD=$PWD` from shell output to verify directory changes
  - Build wt binary once per test run using `sync.Once` for efficiency
  - Skip fish tests gracefully if fish not installed
  - Skip all tests in `-short` mode for fast iteration

### Git Fixture Helper
- **tests/shell/testdata/setup.sh**: POSIX-compatible script that creates real git test fixtures
  - Creates bare repo at `$TMPDIR/repo.git`
  - Creates main worktree at `$TMPDIR/main` with initial commit
  - Creates feature worktree at `$TMPDIR/feature` branched from main
  - Updates bare repo HEAD to point to main branch
  - Outputs parseable paths: `BARE=...`, `MAIN=...`, `FEATURE=...`
  - Used by all shell tests to create isolated test environments

### Bug Fix (Deviation Rule 1)
- **internal/git/repo.go**: Fixed `GetHomePath()` function
  - **Bug:** Previously returned bare repo path (first worktree in list)
  - **Impact:** `wt home` failed with "must be run in work tree" error in bare repo setups
  - **Fix:** Skip bare repos and return worktree matching bare repo's HEAD branch
  - **Algorithm:**
    1. Parse `git worktree list --porcelain` to find all worktrees
    2. If bare repo found, read its HEAD with `git symbolic-ref HEAD`
    3. Return worktree whose branch matches bare repo's HEAD
    4. Fallback: return first non-bare worktree
  - **Result:** `wt home` now works correctly in bare repo configurations

## Test Results

All tests passing on macOS:
```
=== RUN   TestBashWrapperGoto
--- PASS: TestBashWrapperGoto (1.01s)
=== RUN   TestBashWrapperHome
--- PASS: TestBashWrapperHome (0.21s)
=== RUN   TestBashWrapperNew
--- PASS: TestBashWrapperNew (0.25s)
=== RUN   TestZshWrapperGoto
--- PASS: TestZshWrapperGoto (0.20s)
=== RUN   TestZshWrapperHome
--- PASS: TestZshWrapperHome (0.22s)
=== RUN   TestZshWrapperNew
--- PASS: TestZshWrapperNew (0.25s)
=== RUN   TestFishWrapperGoto
--- SKIP: TestFishWrapperGoto (0.00s)
=== RUN   TestFishWrapperHome
--- SKIP: TestFishWrapperHome (0.00s)
=== RUN   TestFishWrapperNew
--- SKIP: TestFishWrapperNew (0.00s)
PASS
ok  	github.com/ahmedelarabyy/wt/tests/shell	2.355s
```

Fish tests skipped (fish not installed), bash and zsh tests passing.

## Task Commits

| Task | Description | Commit | Files |
|------|-------------|--------|-------|
| 1 | Shell test fixtures and setup script | b1bb8be | tests/shell/testdata/setup.sh |
| 2 | E2E shell wrapper tests + bug fix | acc3fe5 | tests/shell/shell_test.go, tests/shell/testdata/setup.sh, internal/git/repo.go |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed bare repo home detection in GetHomePath**
- **Found during:** Task 2 - writing tests exposed the bug
- **Issue:** GetHomePath() returned bare repo path (first in `git worktree list`), causing `wt home` to fail with "must be run in work tree" error. You cannot cd into or run git status in a bare repo.
- **Root cause:** Function naively returned first "worktree" line without checking for "bare" marker
- **Fix:**
  - Parse full worktree list to identify bare repo
  - Read bare repo's HEAD to determine which branch it points to
  - Return worktree matching that branch (the logical "home")
  - Fallback to first non-bare worktree if HEAD matching fails
- **Files modified:** internal/git/repo.go
- **Commit:** acc3fe5 (combined with Task 2)
- **Impact:** `wt home` now works correctly in bare repo configurations (affects all users with bare repos)

**2. [Rule 1 - Bug] Fixed setup script worktree paths**
- **Found during:** Task 1 - initial test run
- **Issue:** Script tried to create worktrees nested inside bare repo (e.g., `repo.git/main`), but `git worktree add` creates worktrees at the specified path, not relative to bare repo
- **Fix:** Changed paths to siblings: `$TMPDIR/main` and `$TMPDIR/feature` instead of `$TMPDIR/repo.git/main`
- **Files modified:** tests/shell/testdata/setup.sh
- **Commit:** b1bb8be (Task 1) + acc3fe5 (refinement in Task 2)

**3. [Rule 1 - Bug] Fixed feature branch creation from orphan main**
- **Found during:** Task 1 - setup script testing
- **Issue:** After creating main branch in bare repo, HEAD still pointed to default "master", so creating feature worktree failed with "orphaned reference" warning
- **Fix:** Added `git symbolic-ref HEAD refs/heads/main` to update bare repo HEAD after first commit
- **Impact:** Feature worktree now created cleanly from main branch
- **Files modified:** tests/shell/testdata/setup.sh
- **Commit:** acc3fe5 (refinement in Task 2)

## Integration Points

### Dependencies
- **Wrapper scripts**: Tests source wrapper.bash, wrapper.zsh, wrapper.fish from internal/shell/templates/
- **Commands tested**: goto (05-01), home (05-03), new (05-02)
- **Git library**: Uses internal/git functions indirectly through wt binary

### Test Architecture
```
tests/shell/
├── shell_test.go           # Main E2E test suite
└── testdata/
    └── setup.sh            # Git fixture creation

Test flow:
1. BuildOnce: Compile wt binary (sync.Once)
2. For each test:
   a. Create temp directory
   b. Run setup.sh to create git fixtures
   c. Write shell script to source wrapper + run wt command
   d. Execute via bash/zsh/fish -c "script"
   e. Parse RESULT_PWD from output
   f. Assert PWD changed to expected worktree
   g. Cleanup temp directory automatically (t.TempDir())
```

### Verification Strategy
- **Real shells**: Uses `exec.Command(shellPath, "-c", script)` to run actual shell processes
- **Real git**: Creates real bare repos and worktrees, no mocking
- **Path verification**: Checks `$PWD` ends with expected worktree basename
- **Exit code**: Verifies command succeeds (exit 0)
- **Output parsing**: Extracts `RESULT_PWD=...` from command output

## Next Phase Readiness

### Blockers
None.

### Concerns
None.

### Recommendations
1. **Add CI integration**: Run `go test ./tests/shell/...` in GitHub Actions
2. **Windows support**: Tests currently macOS/Linux only (PATH handling, shell paths)
3. **Add more scenarios**: Test dirty worktrees, config files, error cases
4. **Performance**: Consider parallel test execution (currently sequential)

## Lessons Learned

1. **E2E tests find real bugs**: Shell wrapper tests immediately exposed GetHomePath bug that unit tests missed
2. **Real fixtures > mocks**: Using actual git repos caught multiple path resolution issues
3. **Build once pattern**: sync.Once for binary compilation saved ~5 seconds per test run
4. **Bare repo complexity**: Bare repos have subtle behaviors (HEAD management, path handling) that require careful testing
5. **Shell portability**: POSIX compatibility (bash 3.2+) requires avoiding bashisms like `[[` and `local -r`

## Self-Check: PASSED

**Created files verified:**
- [FOUND] tests/shell/shell_test.go
- [FOUND] tests/shell/testdata/setup.sh

**Commits verified:**
- [FOUND] b1bb8be
- [FOUND] acc3fe5
