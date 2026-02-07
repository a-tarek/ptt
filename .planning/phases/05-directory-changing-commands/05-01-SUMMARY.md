---
phase: 05
plan: 01
subsystem: directory-navigation
tags: [commands, shell-coordination, worktree-helpers]
requires:
  - 03-02 # ResolveWorktree function
  - 04-01 # Config parsing patterns
provides:
  - goto/home commands with --output-path protocol
  - merge/rebase worktree-to-branch resolution
  - Shared git helpers for worktree placement
affects:
  - 05-02 # wt new will use WorktreePath and helpers
  - 05-03 # wt eject will use WorktreePath
  - 06-01 # Shell wrappers will use --output-path protocol
tech-stack:
  added: []
  patterns:
    - "Hidden persistent flags for shell coordination"
    - "Stderr for confirmations, stdout for machine-readable output"
    - "Auto-detect bare vs regular repo for worktree placement"
key-files:
  created:
    - internal/git/repo.go
    - internal/git/repo_test.go
    - cmd/goto.go
    - cmd/home.go
    - cmd/merge.go
    - cmd/rebase.go
  modified: []
key-decisions:
  - id: output-path-protocol
    decision: "--output-path hidden flag for shell wrapper coordination"
    rationale: "Shell wrappers need machine-readable path on stdout, users get human-friendly messages"
    impact: "All cd commands follow same protocol"
  - id: nested-vs-sibling
    decision: "Auto-detect bare vs regular repo for worktree placement"
    rationale: "Bare repos use nested mode, regular repos use sibling mode - no config needed"
    impact: "Fixes v1.0 compounding bug, intuitive behavior"
  - id: already-there-noop
    decision: "Already-there case prints message to stderr, exits 0"
    rationale: "Not an error condition, just informational"
    impact: "Shell wrapper stays in place, no error noise"
duration: 175s
completed: 2026-02-07
---

# Phase 5 Plan 01: Simple Commands & Git Helpers Summary

**One-liner:** Implemented goto/home/merge/rebase commands with --output-path shell coordination protocol and shared git helpers for bare/regular repo detection and worktree path computation.

## Performance Metrics

- **Total Duration:** 2 min 55 sec
- **Tasks Completed:** 2/2
- **Files Created:** 6
- **Commits:** 2 (atomic per-task)
- **Test Coverage:** Unit tests for WorktreePath logic, integration testing via command help

## What We Accomplished

### 1. Shared Git Helper Functions (internal/git/repo.go)
- **IsBareRepository()** — detects bare vs regular repo via `git rev-parse --is-bare-repository`
- **GetRepoRoot()** — returns repo root (bare or regular)
- **GetHomePath()** — returns first worktree path from `git worktree list`
- **WorktreePath()** — computes target path for new worktrees:
  - Bare repo: nested mode (`/code/wt/staging`)
  - Regular repo: sibling mode (`/code/wt-staging`)
  - Fixes v1.0 compounding bug by always resolving to original repo name
- **CurrentBranch()** — returns current branch name, errors on detached HEAD

### 2. goto Command (cmd/goto.go)
- Switches to a worktree via suffix-match resolution
- `--output-path` flag: outputs only path to stdout (for shell wrapper)
- Without flag: prints confirmation to stdout (direct invocation)
- Always prints confirmation to stderr with branch and dirty status
- Already-there case: prints "Already in {name}" to stderr, exits 0

### 3. home Command (cmd/home.go)
- Switches to main worktree (bare repo root or main checkout)
- Same --output-path protocol as goto
- Handles bare repos by showing "(bare)" instead of branch name
- Already-there case: prints "Already home" to stderr, exits 0

### 4. merge Command (cmd/merge.go)
- Resolves worktree name to branch name via ResolveWorktree
- Runs `git merge <branch>` with stdout/stderr passthrough
- Errors if worktree has no branch (bare/detached)
- Prints "Merging {branch} into current branch..." to stderr

### 5. rebase Command (cmd/rebase.go)
- Identical structure to merge, runs `git rebase <branch>`
- Prints "Rebasing onto {branch}..." to stderr
- Full git conflict handling passthrough

## Task Commits

| Task | Commit | Files | Description |
|------|--------|-------|-------------|
| 1 | 6093ec7 | internal/git/repo.go, internal/git/repo_test.go | Add shared git helper functions |
| 2 | 91ed6ae | cmd/goto.go, cmd/home.go, cmd/merge.go, cmd/rebase.go | Implement goto, home, merge, and rebase commands |

## Files Created

```
internal/git/
├── repo.go           # 5 helper functions (bare detection, paths, branch)
└── repo_test.go      # Unit tests for WorktreePath logic

cmd/
├── goto.go           # goto command with --output-path protocol
├── home.go           # home command with --output-path protocol
├── merge.go          # merge command with worktree resolution
└── rebase.go         # rebase command with worktree resolution
```

## Files Modified

None - all new files.

## Decisions Made

### 1. Shell Coordination Protocol
**Context:** Shell wrapper needs to `cd` based on binary output.

**Decision:** `--output-path` hidden persistent flag on root command. When passed, binary outputs ONLY the target path to stdout. Confirmation messages always go to stderr. Without flag, confirmation also goes to stdout (direct invocation mode).

**Alternatives considered:**
- JSON output format (overkill for single path)
- Separate `--json` flag (more flags to maintain)
- stdout-only output (no visual confirmation for wrapper users)

**Rationale:** Simple, machine-readable for wrapper, human-friendly for direct use. Hidden flag means users never see it in help.

**Impact:** All cd commands (goto, home, new, eject) follow same protocol. Shell wrappers in Phase 6 will be trivial.

### 2. Auto-Detect Bare vs Regular Repo
**Context:** Worktree placement differs for bare repos (nested) vs regular repos (sibling).

**Decision:** Use `IsBareRepository()` to auto-detect mode. No config flag needed.

**Alternatives considered:**
- Config flag for mode (extra complexity, easy to misconfigure)
- Always nested (doesn't work for regular repos)
- Always sibling (doesn't work for bare repos)

**Rationale:** Git already knows if repo is bare. Auto-detection is intuitive and matches user expectations.

**Impact:** Fixes v1.0 compounding bug where `wt new feat-1-a` from `wt-feat-1` created `wt-feat-1-feat-1-a`. Now always resolves to original repo name.

### 3. Already-There No-Op
**Context:** User runs `wt goto feature` when already in `feature`.

**Decision:** Print "Already in feature" to stderr, exit 0 (not an error).

**Alternatives considered:**
- Exit 1 as error (noisy, breaks scripts)
- Silent success (no feedback)
- Warning with exit 0 (confusing signal)

**Rationale:** Not an error condition, just informational. Shell wrapper stays in place, no noise.

**Impact:** Clean UX for repeat commands, shell wrapper doesn't cd, scripts don't break.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

### 1. Test File Import Placement
**Issue:** Initially placed `import "os/exec"` at bottom of test file, causing compilation error.

**Resolution:** Moved import to top of file with other imports.

**Impact:** 30 seconds debugging. Reminder to always place imports at file top.

## Next Phase Readiness

### Blockers
None.

### Dependencies Ready
- ✅ ResolveWorktree (03-02)
- ✅ Config parsing patterns (04-01)
- ✅ Setup action executor (04-02)

### Artifacts Available
- ✅ WorktreePath helper ready for wt new (05-02)
- ✅ IsBareRepository helper ready for wt new (05-02)
- ✅ --output-path protocol ready for shell wrappers (06-01)

### Phase 5 Continuation
**Plan 02 (wt new):** Can immediately use WorktreePath, IsBareRepository, and GetHomePath helpers. Config application already implemented in 04-02.

**Plan 03 (wt eject):** Can use WorktreePath for placement, --output-path protocol for cd coordination.

### Testing Notes
- Unit tests cover WorktreePath logic (bare/sibling modes, duplicate detection)
- Helper functions (IsBareRepository, GetHomePath, CurrentBranch) are thin wrappers - will be tested via integration tests with commands in Phase 7
- Commands tested via help output and compilation

## Self-Check: PASSED

All created files verified on disk:
- ✅ internal/git/repo.go
- ✅ internal/git/repo_test.go
- ✅ cmd/goto.go
- ✅ cmd/home.go
- ✅ cmd/merge.go
- ✅ cmd/rebase.go

All commits verified in git log:
- ✅ 6093ec7 (Task 1)
- ✅ 91ed6ae (Task 2)
