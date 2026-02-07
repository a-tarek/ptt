---
phase: 08-interactive-installer
plan: 01
subsystem: installer
tags: [shell-integration, installer, rc-files, migration, backup, idempotent]

requires:
  - 06-01-shell-wrappers
  - 06-02-shell-completion

provides:
  - wt-install-command
  - rc-file-operations
  - v1-migration-support
  - marker-based-management

affects:
  - 08-02-npm-postinstall

tech-stack:
  added:
    - bufio (user input)
  patterns:
    - marker-block-management
    - backup-and-rollback
    - guided-interactive-walkthrough

key-files:
  created:
    - internal/installer/paths.go
    - internal/installer/rcfile.go
    - internal/installer/rcfile_test.go
    - cmd/install.go
  modified: []

decisions:
  - id: marker-block-style
    choice: conda-certbot-style-markers
    rationale: "Use # >>> wt >>> / # <<< wt <<< markers for clear block identification"
    alternatives: [comment-tags, custom-syntax]

  - id: v1-migration-strategy
    choice: comment-out-with-prefix
    rationale: "Comment out old 'source wt.zsh' lines with [wt v2 migration] prefix for safety"
    alternatives: [delete-lines, leave-both]

  - id: backup-retention
    choice: keep-backup-on-success
    rationale: "Keep .wt-backup file after successful install for user peace of mind"
    alternatives: [delete-backup, ask-user]

  - id: idempotency-check
    choice: marker-presence
    rationale: "Simple HasMarkerBlock() check prevents duplicate installations"
    alternatives: [content-hash, version-tracking]

metrics:
  duration: 4min
  completed: 2026-02-07
---

# Phase 08 Plan 01: Interactive Installer Summary

**One-liner:** Guided `wt install` command with shell detection, RC file modification, v1 migration, and marker-based idempotent management

## What Was Built

Created the `wt install` command—the primary onboarding path for users after `npm install @potato/wt`. The installer provides a guided, interactive walkthrough that safely modifies shell RC files to enable wt v2.

### Core Components

**1. RC File Operations Package (`internal/installer`)**

- **paths.go**: Shell-specific RC file path resolution
  - Handles zsh → `~/.zshrc`, bash → `~/.bashrc` (with macOS `.bash_profile` fallback), fish → `~/.config/fish/config.fish`

- **rcfile.go**: Core RC file manipulation operations
  - Marker block management with conda/certbot-style markers (`# >>> wt >>>` / `# <<< wt <<<`)
  - V1 detection via regex patterns matching `source wt.zsh` and `. wt.zsh` (ignoring commented lines)
  - V1 migration by commenting out old lines with `[wt v2 migration]` prefix
  - Block insertion/removal with proper blank line preservation for file structure
  - Backup/restore operations for safe modification with rollback

- **rcfile_test.go**: Comprehensive unit tests (7 test functions, all passing)
  - Tests for marker detection, v1 pattern matching, line commenting, block insertion/removal
  - Edge case coverage: blocks at start/middle/end of file, empty files, missing blocks

**2. Install Command (`cmd/install.go`)**

Interactive multi-step walkthrough:

1. **Shell Detection**: Auto-detect via `shell.DetectShell()`, confirm with user
2. **RC File Check**: Resolve and display RC file path (with `~` for readability)
3. **Idempotency Check**: Exit cleanly if marker block already present
4. **V1 Migration Detection**: Scan for old `source wt.zsh` lines, show what will be commented
5. **Change Preview**: Display exact marker block to be added, list v1 lines to be commented
6. **User Confirmation**: Proceed/decline with clear manual instructions on decline
7. **Safe Application**: Backup RC file, apply v1 migration if needed, insert marker block, rollback on failure
8. **Success Message**: Instructions to restart terminal or source RC file

**Decline Path**: Users who decline get manual copy-paste instructions with the exact eval line.

**Error Handling**:
- Unsupported shell → manual instructions
- Write failure → automatic rollback to backup
- Missing RC file → creates new file (normal for fresh installs)

## Technical Decisions

### Marker Block Style
**Choice**: Conda/Certbot-style markers (`# >>> wt >>>` / `# <<< wt <<<`)

Standard pattern used by conda, certbot, and other tools that manage shell configuration blocks. Clear visual boundaries, easy to parse, universally recognized.

### V1 Migration Strategy
**Choice**: Comment out with `[wt v2 migration]` prefix

Safety-first approach: old lines aren't deleted, just commented out with a clear migration marker. Users can manually remove them later after confirming v2 works.

**V1 Detection**: Regex patterns match both `source` and dot-source (`.`) syntax, with whitespace tolerance. Commented lines are ignored to avoid false positives.

### Idempotency
**Choice**: Simple marker presence check

`HasMarkerBlock()` checks for `# >>> wt >>>` in content. If found, exit cleanly with "already configured" message. No version tracking or content hashing needed—marker presence is sufficient.

### Backup Retention
**Choice**: Keep backup on success

The `.wt-backup` file is kept after successful installation. Users have peace of mind knowing they can manually revert if needed. Disk space cost is negligible.

### Blank Line Preservation
**Decision**: When removing marker block, leave a blank line in its place

Maintains file structure and visual separation. If block is at start, leave leading blank. If at end, leave trailing blank. If in middle, leave blank for separation between surrounding lines.

## Testing & Verification

**Unit Tests**:
- 7 test functions covering all RC file operations
- Edge cases: blocks at different positions, empty files, v1 patterns with whitespace
- All tests pass

**Integration**:
- `go build ./...` — compiles cleanly
- `go vet ./...` — no issues
- `./wt-bin install --help` — shows help text correctly

**Manual Test Scenarios** (plan specifies, not executed):
- Fresh install to empty RC file
- Idempotent re-run (should exit cleanly)
- V1 migration (should comment out old lines)
- User decline (should show manual instructions)

## Key Files

**Created**:
- `internal/installer/paths.go` (43 lines) — RC file path resolution
- `internal/installer/rcfile.go` (180 lines) — Core RC operations with marker/v1 support
- `internal/installer/rcfile_test.go` (340 lines) — Comprehensive unit tests
- `cmd/install.go` (193 lines) — Interactive install command

**Modified**: None

**Total New Code**: ~756 lines (including tests and comments)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] RemoveMarkerBlock blank line handling**

- **Found during**: Task 1 unit test execution
- **Issue**: `RemoveMarkerBlock()` wasn't preserving blank lines correctly. When removing a marker block, the function should leave a blank line in its place to maintain file structure, but initial implementation joined lines without preserving the gap.
- **Fix**: Updated logic to add an empty string to the result array when encountering `MarkerBegin`, which becomes a blank line when joined. This maintains visual separation in the RC file.
- **Files modified**: `internal/installer/rcfile.go`
- **Commit**: 6b9e17c (included in task commit)
- **Test evidence**: All `TestRemoveMarkerBlock` cases now pass, including edge cases for blocks at start/middle/end

## Task Commits

| Task | Description | Commit | Files |
|------|-------------|--------|-------|
| 1 | Create installer package for RC file operations | 6b9e17c | internal/installer/paths.go, rcfile.go, rcfile_test.go |
| 2 | Create wt install command with guided walkthrough | 17fc7ae | cmd/install.go |

## Next Phase Readiness

**Ready for**: 08-02 npm postinstall integration

The installer command is complete and tested. Next plan can integrate it into the npm package lifecycle:
- Add postinstall hook that prompts user to run `wt install`
- Or auto-run installer with appropriate prompts
- Document in npm package README

**Provides**:
- `wt install` command (user-facing, interactive)
- RC file operations (internal/installer package)
- V1 migration support (automatic detection and commenting)
- Idempotent behavior (safe to run multiple times)

**Dependencies satisfied**:
- Shell detection from 06-01 (shell.DetectShell)
- Shell-init command from 06-01 (eval line integration)
- Shell wrapper templates from 06-01 (referenced in marker block)

## Must-Haves Status

All must-haves satisfied:

✅ User can run `wt install` and see detected shell with confirmation prompt
✅ Installer shows what will be added to rc file before modifying
✅ Running installer twice does not duplicate entries (idempotent)
✅ User declining installation gets manual copy-paste instructions
✅ RC file modifications use marker blocks (>>> wt >>> / <<< wt <<<)
✅ Existing wt v1 entries (source wt.zsh) are commented out during install
✅ RC file is backed up before modification
✅ Partial modifications are rolled back on failure

All artifacts meet minimum line requirements:
- internal/installer/rcfile.go: 180 lines (min: 80) ✅
- internal/installer/rcfile_test.go: 340 lines (min: 60) ✅
- internal/installer/paths.go: 43 lines (min: 20) ✅
- cmd/install.go: 193 lines (min: 80) ✅

All key links verified:
- cmd/install.go → internal/installer/rcfile.go via function calls ✅
- cmd/install.go → internal/shell/detect.go via shell.DetectShell() ✅

## Self-Check: PASSED

**Created files verification**:
```bash
[ -f "internal/installer/paths.go" ] && echo "FOUND"
[ -f "internal/installer/rcfile.go" ] && echo "FOUND"
[ -f "internal/installer/rcfile_test.go" ] && echo "FOUND"
[ -f "cmd/install.go" ] && echo "FOUND"
```
All files exist ✅

**Commits verification**:
```bash
git log --oneline | grep -E "(6b9e17c|17fc7ae)"
```
Both commits present in history ✅

---

## Self-Check Results

**Files Created**: ✅ All 4 files exist
- internal/installer/paths.go ✅
- internal/installer/rcfile.go ✅
- internal/installer/rcfile_test.go ✅
- cmd/install.go ✅

**Commits**: ✅ Both commits present in history
- 6b9e17c: feat(08-01): create installer package for RC file operations ✅
- 17fc7ae: feat(08-01): create wt install command with guided walkthrough ✅

**Self-Check Status**: ✅ PASSED
