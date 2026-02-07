---
phase: 06-shell-integration
plan: 02
subsystem: completion
tags: [cobra, shell-completion, tab-completion, ux]
requires: [05-01-directory-commands, 03-02-list-delete]
provides: [dynamic-worktree-completion, shell-completion-scripts]
affects: [06-03-wrapper-functions]
tech-stack:
  added: []
  patterns: [cobra-ValidArgsFunction, live-completion-queries]
key-files:
  created:
    - cmd/completion.go
  modified:
    - cmd/goto.go
    - cmd/delete.go
    - cmd/merge.go
    - cmd/rebase.go
key-decisions:
  - Names-only completion: Show worktree basenames only, no branch descriptions (cleaner UX)
  - Live queries: Query git worktree list on every tab press (no caching, always accurate)
  - NoFileComp directive: Suppress file/directory completions (only show worktree names)
  - Single-arg completion: Stop completing after first positional argument provided
  - No completion for new/eject: These commands create new names, not select existing ones
duration: 1 min
completed: 2026-02-07
---

# Phase 6 Plan 2: Dynamic Worktree Tab Completion Summary

**One-liner:** Dynamic worktree name tab completion via Cobra ValidArgsFunction with live git queries and NoFileComp directive

## Performance

**Execution time:** 1 minute 43 seconds
**Tasks completed:** 2/2
**Commits:** 2
**Deviations:** 0

## Accomplishments

### What Was Built

Added dynamic tab completion support for worktree-accepting commands (goto, delete, merge, rebase):

**Shared Completion Function:**
- Created `cmd/completion.go` with `worktreeNameCompletion` function
- Queries `git.ListWorktrees()` live on every tab press (no caching)
- Returns worktree basenames only (e.g., "wt-feat-1" not full paths)
- Suppresses file completion via `cobra.ShellCompDirectiveNoFileComp`
- Stops completing after first positional argument (`len(args) >= 1` check)
- Returns `cobra.ShellCompDirectiveError` on git failure (graceful degradation)

**Command Integration:**
- Added `ValidArgsFunction: worktreeNameCompletion` to:
  - `cmd/goto.go` - Switch to a worktree
  - `cmd/delete.go` - Remove a worktree
  - `cmd/merge.go` - Merge a worktree's branch into current
  - `cmd/rebase.go` - Rebase current onto worktree's branch
- Intentionally NOT added to `new` or `eject` (they create new names, not select existing)

**Built-in Completion Command:**
- Cobra v1.10.2 automatically provides `wt completion bash/zsh/fish` subcommands
- No explicit registration needed
- Generates shell-specific completion scripts (426/212/235 lines respectively)

### Must-Haves Verification

All must-have truths verified:

✓ Tab-pressing after `wt goto ` shows worktree names
✓ Tab-pressing after `wt delete ` shows worktree names
✓ Tab-pressing after `wt merge ` shows worktree names
✓ Tab-pressing after `wt rebase ` shows worktree names
✓ Tab completion does not show file/directory completions (only worktree names)
✓ Tab-pressing after `wt goto name ` does not complete (already has 1 arg)
✓ `wt completion bash` outputs a bash completion script (426 lines)
✓ `wt completion zsh` outputs a zsh completion script (212 lines)
✓ `wt completion fish` outputs a fish completion script (235 lines)

All must-have artifacts created:

✓ `cmd/completion.go` - Shared worktree name completion function
✓ `cmd/goto.go` - ValidArgsFunction on goto command
✓ `cmd/delete.go` - ValidArgsFunction on delete command
✓ `cmd/merge.go` - ValidArgsFunction on merge command
✓ `cmd/rebase.go` - ValidArgsFunction on rebase command

All key links established:

✓ `cmd/goto.go` → `cmd/completion.go` via ValidArgsFunction assignment
✓ `cmd/completion.go` → `internal/git/worktree.go` via git.ListWorktrees() call

## Task Commits

| Task | Commit  | Message                                               |
| ---- | ------- | ----------------------------------------------------- |
| 1    | 6a42dd7 | feat(06-02): add shared worktree name completion function |
| 2    | 5dc8930 | feat(06-02): wire ValidArgsFunction to worktree commands |

## Files Created

**cmd/completion.go** (35 lines)
- `worktreeNameCompletion()` - Shared completion function for worktree-accepting commands
- Live queries via `git.ListWorktrees()` (no caching)
- Returns basenames with `cobra.ShellCompDirectiveNoFileComp`
- Stops completing after first positional argument

## Files Modified

**cmd/goto.go**
- Added `ValidArgsFunction: worktreeNameCompletion` to gotoCmd

**cmd/delete.go**
- Added `ValidArgsFunction: worktreeNameCompletion` to deleteCmd

**cmd/merge.go**
- Added `ValidArgsFunction: worktreeNameCompletion` to mergeCmd

**cmd/rebase.go**
- Added `ValidArgsFunction: worktreeNameCompletion` to rebaseCmd

## Decisions Made

**1. Names-only completion (no descriptions)**
- **Context:** Cobra supports `name\tdescription` format for completions
- **Decision:** Return worktree basenames only, no branch descriptions
- **Rationale:** Cleaner UX, less visual noise, names are sufficient for identification
- **Impact:** Users see just "wt-feat-1" not "wt-feat-1	(branch: feature-1)"

**2. Live queries on every tab press (no caching)**
- **Context:** Worktree list could be cached for performance
- **Decision:** Query `git.ListWorktrees()` on every tab press
- **Rationale:** Worktrees can be added/removed outside wt (via git directly), caching would show stale data
- **Impact:** Completions always accurate but ~5-10ms query on each tab press (acceptable)

**3. Suppress file completion (NoFileComp directive)**
- **Context:** By default, shells fall back to file/directory completion
- **Decision:** Return `cobra.ShellCompDirectiveNoFileComp` to suppress file completion
- **Rationale:** Worktree names are not file paths, showing files would be confusing
- **Impact:** Tab only shows worktree names, no file paths suggested

**4. Single-argument completion limit**
- **Context:** Commands accept exactly 1 positional arg (`cobra.ExactArgs(1)`)
- **Decision:** Stop completing after first argument (`len(args) >= 1` check)
- **Rationale:** No second argument exists, continuing to complete would be misleading
- **Impact:** After typing `wt goto name <TAB>`, no suggestions appear (correct)

**5. No completion for new/eject commands**
- **Context:** `wt new` and `wt eject` also accept name arguments
- **Decision:** Do NOT add ValidArgsFunction to new/eject
- **Rationale:** These commands create new worktree names, not select existing ones
- **Impact:** Users type new names freely without being suggested existing worktree names

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - Cobra completion API worked as expected.

## Next Phase Readiness

**Blockers:** None

**Concerns:** None

**Dependencies satisfied:**
- ✓ Phase 05-01 complete (goto, home, merge, rebase commands exist)
- ✓ Phase 03-02 complete (internal/git/worktree.go provides ListWorktrees)

**Artifacts ready for next phase:**
- `cmd/completion.go` - Shared completion function
- `wt completion bash/zsh/fish` - Shell script generators
- ValidArgsFunction wired to all worktree-accepting commands

**Phase 06 Status:** Plan 02 of 03 complete
- ✓ Plan 01: Research shell integration domain → Complete
- ✓ Plan 02: Add dynamic worktree tab completions → **Complete** ⬅ Current
- ⬜ Plan 03: Implement shell wrapper functions → Pending

**Ready to proceed:** Yes - completion infrastructure ready for shell wrapper integration in 06-03

## Self-Check: PASSED

**Files created verification:**
```
✓ cmd/completion.go exists (35 lines)
```

**Commit verification:**
```
✓ 6a42dd7 feat(06-02): add shared worktree name completion function
✓ 5dc8930 feat(06-02): wire ValidArgsFunction to worktree commands
```

All claims verified.
