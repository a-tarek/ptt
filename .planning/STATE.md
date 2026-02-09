# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-09)

**Core value:** A single `ptt` command that works in any shell on any platform with full autocompletion
**Current focus:** v2.0 Bare Repo + cd Rename -- Phase 17 complete (mk-bare-repo command)

## Current Position

Phase: 18 of 19 (Adopt Smart Init)
Plan: 3 of 3 in current phase
Status: Phase complete
Last activity: 2026-02-09 -- Completed 18-03-PLAN.md

Progress: [████████████████░░░░] 87% (40/~46 plans -- 40 complete)

## Performance Metrics

**Previous milestones:**
- v1.0 Documentation: 6 plans, ~2 min/plan
- v2.0 Go Rewrite: 20 plans, ~3.2 min/plan
- v2.0 Rebrand: 6 plans, ~3.5 min/plan
- v2.0 Bare Repo Infrastructure: 2 plans, ~4.5 min/plan
- v2.0 cd Rename: 1 plan, ~6 min/plan
- v2.0 mk-bare-repo Command: 2 plans, ~2 min/plan
- v2.0 Adopt Smart Init: 3 plans, ~3 min/plan (complete)
- Total: 40 plans executed in ~2.05 hours

## Accumulated Context

### Decisions

Previous decisions logged in PROJECT.md Key Decisions table.

New decisions for this milestone:
- **Rename go -> cd**: `cd` is more intuitive; `go` removed entirely (no alias)
- **Clean break, no aliases** (16-01): Removed go/goto/home completely for simplicity, no backward compat
- **Bare repo support**: Nested worktrees inside bare repo avoids cluttering parent directory
- **mk-bare-repo as clone-from-remote**: Safer than in-place restructure -- user verifies then deletes old
- **.pttconfig at bare root**: Project-level config shared by all worktrees, not per-worktree
- **BareRepoRoot() as foundation**: All bare repo awareness lives in `internal/git/repo.go`
- **git rev-parse --git-common-dir for detection** (15-01): Most reliable way to detect ptt bare repos
- **Backward compatibility in WorktreePath()** (15-01): Support both ptt and standard bare repos
- **Integration tests with real git** (15-01): Use actual git commands, not mocks, for reliability
- **ConfigRoot() for config, GetHomePath() for file operations** (15-02): Separation ensures .pttconfig/ is shared across worktrees while files are copied from current location
- **Clone-from-remote conversion** (17-01): mk-bare-repo clones remote into .bare/ rather than restructure in-place for safety
- **Target directory naming** (17-01): <repo>-bare/ as sibling allows user verification before deleting original
- **Initial worktree creation** (17-01): Always create worktree for default branch (main or master) for immediate usability
- **mk-bare-repo documentation positioning** (17-02): Placed between mk and cd commands to group create commands together
- **Error messages documented** (17-02): Self-diagnosis table helps users troubleshoot without support
- **Package name 'initcmd'** (18-01): Avoid Go keyword 'init' for package name
- **RepoType detection order** (18-01): Check IsInsideGitRepo → BareRepoRoot → IsBareRepository → default to normal
- **IsBareFromWorktree flag** (18-01): Distinguishes calling init from container root vs from within a worktree
- **ProgressCallback export** (18-01): Exported type needed by Plan 18-02 for restructure/adopt/repair operations
- **Staging directory for untracked files** (18-02): Normal repo restructure preserves untracked files via _ptt_staging/
- **Cleanup stack pattern** (18-02): All transformations build rollback functions, executed in reverse on failure
- **Feature branch sanitization** (18-02): Replace "/" with "-" in branch names for safe directory names
- **IsDirty() ignores untracked files** (18-03): Only actual changes (M/A/D) trigger dirty state, untracked files allowed
- **RepairPttRepo() handles .pttconfig** (18-03): .pttconfig creation is a valid repair item for ptt bare repos
- **Default branch from remote HEAD** (18-03): Normal repos detect default from remote HEAD, not current checkout

### Pending Todos

None.

### Blockers/Concerns

**NPM_TOKEN secret required for first release:**
- Release workflow needs NPM_TOKEN configured in repository settings
- Recommendation: Test with beta tag (v2.0.0-beta.1) before v2.0.0

## Session Continuity

Last session: 2026-02-09
Stopped at: Completed 18-03-PLAN.md -- Phase 18 complete (Adopt Smart Init)
Resume file: None

---
*State initialized: 2026-02-07*
*Updated: 2026-02-09 -- Phase 18 complete (Smart Init command with comprehensive tests and bug fixes)*
