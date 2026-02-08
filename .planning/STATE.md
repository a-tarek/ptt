# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-08)

**Core value:** A single `ptt` command that works in any shell on any platform with full autocompletion
**Current focus:** v2.0 Pre-Release — Rebrand + Command Restructure

## Current Position

Phase: Not started (defining requirements)
Plan: —
Status: Defining requirements
Last activity: 2026-02-08 — v2.0 rebrand planning started

Progress: Phases 1-10 complete, rebrand phases pending

## Performance Metrics

**Velocity:**
- Total plans completed: 20
- Average duration: ~2.6 min per plan
- Total execution time: ~0.87 hours

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
| 9. Polish & Testing | 4/4 ✓ | 14 min | 3.5 min |
| 10. UAT Gap Closure | 2/2 ✓ | 6.2 min | 3.1 min |

**Recent Trend:**
- Phase 10 (UAT Gap Closure) COMPLETE ✓
- Plan 10-01 complete: Stdout leak fix + --run flag — clean stderr/stdout separation for shell wrapper (4 min)
- Plan 10-02 complete: Segment-aware fuzzy matching — scoring system for worktree name resolution (2.2 min)
- Phase 9 (Polish & Testing) COMPLETE ✓
- Plan 09-01 complete: Enhanced error messages — fuzzy matching, color output, help footer (3 min)
- Plan 09-02 complete: CI/CD pipeline — GitHub Actions with test matrix + release automation (1.6 min)
- Plan 09-03 complete: Shell E2E tests — bash/zsh/fish wrapper tests with real git repos (7.3 min)
- Plan 09-04 complete: README rewrite — 635-line v2 documentation (2.2 min)

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

**Phase 9 Decisions:**
- **CI matrix OS**: Test on ubuntu-latest and macos-latest only - these are target platforms, Windows deferred (09-02)
- **Coverage tracking**: Display coverage in output but don't enforce thresholds - visibility without build failures (09-02)
- **Go version pinning**: Use exact version from go.mod (1.25.7) for predictable CI environment (09-02)
- **Lean CI philosophy**: Only go vet, no golangci-lint - keep CI fast and simple (09-02)
- **Single-command release**: Full automation on tag push - git tag v2.0.0 triggers complete pipeline (09-02)
- **Real git fixtures**: Use real git repos with worktrees for E2E tests (no mocking) - highest confidence shell integration works (09-03)
- **Build once pattern**: Build wt binary once per test run using sync.Once - 6x faster test execution (09-03)
- **Skip missing shells**: Fish tests skip gracefully if not installed - allows tests to pass on systems without all shells (09-03)
- **Short mode support**: All tests check testing.Short() to allow fast iteration with go test -short (09-03)

**Phase 10 Decisions:**
- **Status messages to stderr**: Progress messages (Copied, Symlinked, Running) use fmt.Fprintf(os.Stderr) to prevent shell wrapper stdout capture interference (10-01)
- **Run command output to stderr**: cmd.Stdout = os.Stderr ensures run-action output doesn't leak into cd path (10-01)
- **--run flag for convenience**: Enables wt new feature --run 'npm install' without .wtconfig file for AI agents and quick workflows (10-01)
- **Segment-aware scoring**: Replace pure Levenshtein with scoring system (0-100) for fuzzy matching - handles wt-prefix pattern correctly (10-02)
- **Export FindClosestMatch**: Export function for direct unit testing - enables comprehensive test coverage (10-02)
- **Case-insensitive matching**: Convert to lowercase for matching, preserve original case in results - better UX (10-02)

### Pending Todos

None yet.

### Blockers/Concerns

**NPM_TOKEN secret required for first release:**
- Release workflow needs NPM_TOKEN configured in repository settings
- Without it, automated npm publish will fail
- Documented in .github/workflows/release.yml comments
- Recommendation: Test with beta tag (v2.0.0-beta.1) before v2.0.0

## Session Continuity

Last session: 2026-02-08
Stopped at: Phase 10 complete — all UAT gaps closed, ready for milestone audit
Resume file: None

---
*State initialized: 2026-02-07*
*v1.0 complete, v2.0 roadmap ready*
