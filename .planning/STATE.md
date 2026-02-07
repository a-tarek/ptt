# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-07)

**Core value:** A single `wt` command that works in any shell on any platform with full autocompletion
**Current focus:** Phase 5 - Directory-Changing Commands (goto, home, merge, rebase, new, eject)

## Current Position

Phase: 5 of 9 (Directory-Changing Commands)
Plan: 2 of 3 complete
Status: In progress
Last activity: 2026-02-07 — Completed 05-02-PLAN.md (wt new command)

Progress: [████░░░░░░] 47% (4/9 phases complete + 2/3 of phase 5)

## Performance Metrics

**Velocity:**
- Total plans completed: 12
- Average duration: ~2.5 min per plan
- Total execution time: ~0.62 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1. Internal Documentation | 3 | - | ~2 min |
| 2. User-Facing Documentation | 3 | - | ~2 min |
| 3. Core Go Binary Foundation | 2/2 ✓ | 7 min | 3.5 min |
| 4. Configuration System | 2/2 ✓ | 8 min | 4 min |
| 5. Directory-Changing Commands | 2/3 | 5 min | 2.5 min |

**Recent Trend:**
- Phase 5 (Directory-Changing Commands) in progress (2/3 complete)
- Plan 05-01 complete: goto/home/merge/rebase + shared git helpers
- Plan 05-02 complete: wt new command with config integration
- Fixed bare repo detection bug (WorktreePath now checks home path)
- Ready for Plan 05-03 (eject command)

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- **Go over Node.js/bash**: Fast startup (~5ms vs ~150ms), single binary, cobra completions for free (✓ Implemented - 03-01)
- **npm for distribution**: Download tracking, familiar install (npx), cross-platform binary delivery (Pending implementation)
- **Shell wrappers for cd**: Subprocess can't change parent directory — standard pattern (zoxide, nvm) (Pending implementation)
- **Scoped npm package**: "wt" taken on npm; @scope/wt guarantees availability, CLI stays `wt` (Pending implementation)
- **Keep wt.zsh as legacy**: Existing users shouldn't break, gradual migration (Pending implementation)
- **Port only, no new features**: Clean port reduces risk, new features come after v2.0 (Pending implementation)

**Phase 3 Decisions:**
- **Dirty indicator (~)**: Used tilde for dirty status - widely understood, works without Unicode issues (03-01)
- **.wtconfig location**: Created in current directory (not repo root) - supports per-directory configs (03-01) [SUPERSEDED by 04-01]
- **Silent success**: init command exits silently on success - follows git-style UX patterns (03-01)
- **Suffix matching resolution**: Worktree names resolve via suffix match (e.g., "staging" matches "repo-staging") - user-friendly (03-02)
- **Confirmation only for dirty**: Clean worktrees delete silently, dirty prompt for confirmation - balances safety with convenience (03-02)
- **Conservative branch deletion**: --branch flag required to delete branch, not default - prevents accidental branch loss (03-02)

**Phase 4 Decisions:**
- **Config files at repo root**: .wtconfig and .wtconfig-* live at repo root, not cwd - matches v1.0 and supports multi-directory repos (04-01)
- **SplitN for parsing**: Use SplitN(line, ' ', 2) not Split() to preserve spaces in run commands (04-01)
- **Collected error reporting**: ValidateActions reports all errors at once, not fail-on-first - better UX (04-01)
- **Bare name resolution**: Bare names like "ci" resolve to .wtconfig-ci, paths with "/" treated as exact - intuitive (04-01)
- **Type-grouped flag order**: Inline flags ordered by type (copy, symlink, run) due to Cobra limitation - documented (04-01)
- **Use otiai10/copy library**: External dependency chosen over hand-rolled implementation for reliability (04-02)
- **Absolute symlink paths**: Symlinks use absolute source paths for consistent resolution regardless of working directory (04-02)
- **Rollback with fallback**: git worktree remove with fallback to os.RemoveAll ensures cleanup always attempted (04-02)

**Phase 5 Decisions:**
- **--output-path protocol**: Hidden persistent flag outputs only path to stdout for shell wrapper coordination - confirmations always go to stderr (05-01)
- **Auto-detect bare vs regular**: Use IsBareRepository() to determine nested vs sibling mode - no config needed, fixes v1.0 compounding bug (05-01)
- **Already-there no-op**: Print message to stderr and exit 0 when already in target worktree - not an error condition (05-01)
- **Config merge on top**: --copy/--symlink flags merge with file-based config, apply independently of --skip-config (05-02)
- **Silent config skip**: Missing .wtconfig is silently skipped, not an error - only projects that need it create it (05-02)
- **Bare detection via home path**: WorktreePath checks if home path is bare, not current directory - when in worktree of bare repo, IsBareRepository() returns false (05-02)

### Pending Todos

None yet.

### Blockers/Concerns

None yet — v2.0 roadmap created, ready to plan Phase 3.

## Session Continuity

Last session: 2026-02-07
Stopped at: Completed 05-02-PLAN.md (wt new command) - Phase 5 in progress (2/3 complete)
Resume file: None

---
*State initialized: 2026-02-07*
*v1.0 complete, v2.0 roadmap ready*
