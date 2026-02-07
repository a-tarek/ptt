# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-07)

**Core value:** A single `wt` command that works in any shell on any platform with full autocompletion
**Current focus:** Phase 8 - Interactive Installer

## Current Position

Phase: 8 of 9 (Interactive Installer)
Plan: 2 of 2 complete
Status: Phase complete
Last activity: 2026-02-07 — Completed 08-02-PLAN.md

Progress: [████████░] 89% (16/18 total plans complete)

## Performance Metrics

**Velocity:**
- Total plans completed: 16
- Average duration: ~2.3 min per plan
- Total execution time: ~0.73 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1. Internal Documentation | 3 | - | ~2 min |
| 2. User-Facing Documentation | 3 | - | ~2 min |
| 3. Core Go Binary Foundation | 2/2 ✓ | 7 min | 3.5 min |
| 4. Configuration System | 2/2 ✓ | 8 min | 4 min |
| 5. Directory-Changing Commands | 3/3 ✓ | 9 min | 3 min |
| 6. Shell Integration | 2/2 ✓ | 3 min | 1.5 min |
| 7. npm Distribution | 2/2 ✓ | 7 min | 3.5 min |
| 8. Interactive Installer | 2/2 ✓ | 6 min | 3 min |

**Recent Trend:**
- Phase 8 (Interactive Installer) COMPLETE ✓
- Plan 08-01 complete: wt install command — guided walkthrough, RC file operations, v1 migration (4 min)
- Plan 08-02 complete: wt uninstall command — clean removal path with confirmation flow (2 min)
- Marker-based idempotent installation with backup/rollback
- Conda/certbot-style markers for clear block management
- V1 detection and safe migration (comment out old source lines)
- Paired install/uninstall commands for complete lifecycle
- Ready for Phase 9 (Remaining Features)

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- **Go over Node.js/bash**: Fast startup (~5ms vs ~150ms), single binary, cobra completions for free (✓ Implemented - 03-01)
- **npm for distribution**: Download tracking, familiar install (npx), cross-platform binary delivery (✓ In progress - 07-01)
- **Shell wrappers for cd**: Subprocess can't change parent directory — standard pattern (zoxide, nvm) (✓ Implemented - 06-01)
- **Scoped npm package**: "wt" taken on npm; @scope/wt guarantees availability, CLI stays `wt` (✓ Implemented - 07-01 with @potato scope)
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
- **Stash conflict non-fatal**: When stash pop has merge conflicts, print warning but don't fail - stash was still popped, user can resolve (05-03)
- **Fallback branch detection**: Home worktree (srcRoot == homePath) uses main/master, non-home uses directory suffix (05-03)
- **Multi-step rollback**: Each eject step has rollback path - pop stash, checkout back, remove worktree (05-03)

**Phase 6 Decisions:**
- **POSIX-compatible bash wrapper**: Use [ ] not [[ ]] for bash 3.2+ compatibility (macOS default) - slightly more verbose but maximum compatibility (06-01)
- **Identical bash and zsh wrappers**: POSIX constructs work in both shells - easier to maintain, reduced testing surface (06-01)
- **Hidden shell-init command**: Set Hidden: true in cobra - plumbing command for rc files, follows git-style UX (06-01)
- **Route all commands through wrapper**: Simpler mental model (like zoxide) - non-cd commands produce no output with --output-path, so wrapper is pass-through (06-01)
- **Names-only completion**: Show worktree basenames only, no branch descriptions - cleaner UX, less visual noise (06-02)
- **Live completion queries**: Query git worktree list on every tab press (no caching) - always accurate, ~5-10ms cost acceptable (06-02)
- **NoFileComp directive**: Suppress file/directory completions for worktree args - only show worktree names (06-02)
- **Single-arg completion limit**: Stop completing after first positional argument - no second arg exists (06-02)
- **No completion for new/eject**: These commands create new names, not select existing ones (06-02)

**Phase 7 Decisions:**
- **@potato npm scope**: User chose @potato for all packages - guarantees npm availability, CLI remains `wt` (07-01)
- **Platform naming**: Package names use Go arch (amd64), cpu field uses npm arch (x64) - follows ecosystem conventions (07-01)
- **Node.js wrapper mapping**: Wrapper maps Node.js process.arch to Go arch names for package resolution (07-01)
- **.gitignore exception**: Added !npm/bin/wt to allow wrapper script commit while blocking Go binary (07-01)
- **goreleaser path mapping**: Build script handles _v8.0 suffix for arm64 and _v1 for amd64 builds (07-02)
- **jq-free version updates**: Publish script uses node -e for JSON manipulation, no jq dependency (07-02)
- **Platform-first publishing**: Platform packages published before main package to satisfy dependency order (07-02)

**Phase 8 Decisions:**
- **Marker-block style**: Conda/certbot-style markers (# >>> wt >>> / # <<< wt <<<) for clear block identification (08-01)
- **V1 migration strategy**: Comment out old 'source wt.zsh' lines with [wt v2 migration] prefix for safety (08-01)
- **Backup retention**: Keep .wt-backup file after successful install for user peace of mind (08-01)
- **Idempotency check**: Simple HasMarkerBlock() marker presence check prevents duplicate installations (08-01)
- **No v1 uncommenting on uninstall**: User can manually uncomment v1 lines if reverting - automatic uncommenting could be fragile (08-02)
- **Uninstall only cleans rc file**: Does not run npm uninstall (self-destructive mid-execution), prints instructions instead (08-02)

### Pending Todos

None yet.

### Blockers/Concerns

None yet — v2.0 roadmap created, ready to plan Phase 3.

## Session Continuity

Last session: 2026-02-07
Stopped at: Completed 08-02-PLAN.md (Interactive installer) - Phase 8 complete
Resume file: None

---
*State initialized: 2026-02-07*
*v1.0 complete, v2.0 roadmap ready*
