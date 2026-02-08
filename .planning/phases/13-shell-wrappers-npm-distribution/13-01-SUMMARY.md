---
phase: 13-shell-wrappers-npm-distribution
plan: 01
subsystem: shell-integration
tags: [rebrand, shell, installer, wrappers, bash, zsh, fish]
requires: [12-02]
provides:
  - ptt() shell function in all three shells (bash, zsh, fish)
  - __PTT_BIN__ placeholder in shell wrappers
  - RC file markers using >>> ptt >>> format
  - Eval lines referencing ptt shell-init
  - Backup suffix .ptt-backup
affects: [13-02, 14-01]
tech-stack:
  added: []
  patterns: [shell-function-interception, cd-command-delegation]
key-files:
  created: []
  modified:
    - internal/shell/templates/wrapper.bash
    - internal/shell/templates/wrapper.zsh
    - internal/shell/templates/wrapper.fish
    - internal/shell/embed.go
    - internal/installer/rcfile.go
    - internal/installer/rcfile_test.go
    - tests/shell/shell_test.go
key-decisions: []
duration: 203s
completed: 2026-02-08
---

# Phase 13 Plan 01: Shell Wrapper Rebrand Summary

**One-liner:** Rebrand shell wrappers from wt to ptt, updating function names, RC markers, and all command interception patterns

## Performance

- **Duration:** 203 seconds (~3.4 minutes)
- **Start time:** 2026-02-08T21:20:14Z
- **End time:** 2026-02-08T21:23:38Z
- **Tasks completed:** 2/2
- **Files modified:** 7

## What We Accomplished

Completed full rebrand of shell integration layer from `wt` to `ptt`:

1. **Shell wrapper templates** - All three shell wrappers (bash, zsh, fish) now define `ptt()` function instead of `wt()`
2. **Command interception** - Updated case statements to intercept both new primary commands (`go`, `mk`) and backward-compatible aliases (`goto`, `home`, `new`)
3. **Binary placeholder** - Changed from `__WT_BIN__` to `__PTT_BIN__` in all templates and embed.go
4. **RC file markers** - Updated from `# >>> wt >>>` / `# <<< wt <<<` to `# >>> ptt >>>` / `# <<< ptt <<<`
5. **Eval lines** - Changed from `eval "$(wt shell-init)"` to `eval "$(ptt shell-init)"` and `wt shell-init | source` to `ptt shell-init | source`
6. **Backup suffix** - Updated from `.wt-backup` to `.ptt-backup`
7. **Migration comments** - Changed from `[wt v2 migration]` to `[ptt migration]`
8. **Test suite** - Updated all tests to use new branding while maintaining full test coverage

## Task Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1 | 8ca91c0 | Rebrand shell wrapper templates to ptt |
| 2 | ddb34bf | Rebrand RC file management and update all tests |

## Files Created/Modified

### Modified Files

**internal/shell/templates/wrapper.bash**
- Changed function name from `wt()` to `ptt()`
- Updated case statement: `goto|home|new|eject` → `go|goto|home|mk|new|eject`
- Replaced all `__WT_BIN__` with `__PTT_BIN__` (3 occurrences)

**internal/shell/templates/wrapper.zsh**
- Changed function name from `wt()` to `ptt()`
- Updated case statement: `goto|home|new|eject` → `go|goto|home|mk|new|eject`
- Replaced all `__WT_BIN__` with `__PTT_BIN__` (3 occurrences)

**internal/shell/templates/wrapper.fish**
- Changed function name from `function wt` to `function ptt`
- Updated case statement: `goto home new eject` → `go goto home mk new eject`
- Replaced all `__WT_BIN__` with `__PTT_BIN__` (3 occurrences)

**internal/shell/embed.go**
- Updated `binPlaceholder` constant from `__WT_BIN__` to `__PTT_BIN__`

**internal/installer/rcfile.go**
- Updated `MarkerBegin` from `"# >>> wt >>>"` to `"# >>> ptt >>>"`
- Updated `MarkerEnd` from `"# <<< wt <<<"` to `"# <<< ptt <<<"`
- Updated `EvalLine` fish case from `wt shell-init | source` to `ptt shell-init | source`
- Updated `EvalLine` default case from `eval "$(wt shell-init)"` to `eval "$(ptt shell-init)"`
- Updated `MarkerBlock` managed comment from `Managed by wt installer` to `Managed by ptt installer`
- Updated `CommentOutLines` prefix from `# [wt v2 migration] ` to `# [ptt migration] `
- Updated `BackupFile` suffix from `.wt-backup` to `.ptt-backup`
- Preserved v1 detection patterns (`v1SourcePattern`, `v1DotPattern`) that still match `wt.zsh` for backward compatibility

**internal/installer/rcfile_test.go**
- Replaced all `# >>> wt >>>` with `# >>> ptt >>>`
- Replaced all `# <<< wt <<<` with `# <<< ptt <<<`
- Replaced all `eval "$(wt shell-init)"` with `eval "$(ptt shell-init)"`
- Replaced all `wt shell-init | source` with `ptt shell-init | source`
- Replaced all `Managed by wt installer` with `Managed by ptt installer`
- Replaced all `# [wt v2 migration] ` with `# [ptt migration] ` (with trailing space)

**tests/shell/shell_test.go**
- Renamed `buildWtBinary` to `buildPttBinary`
- Renamed `wtBinaryPath` to `pttBinaryPath`
- Changed temp dir prefix from `wt-test-*` to `ptt-test-*`
- Updated binary output path from `wt` to `ptt`
- Updated all test scripts to call `ptt` command instead of `wt`
- Updated error messages from `wt list` to `ptt list`

## Decisions Made

None - This was a pure rebrand following established conventions from Phase 11 and 12.

## Deviations from Plan

None - Plan executed exactly as written.

## Issues Encountered

None - All tests passed on first attempt after correcting test expectations for migration comment spacing.

## Next Phase Readiness

**Phase 13 Plan 02 (npm package rebrand):** Ready to proceed
- Shell wrapper rebrand complete
- All shell integration tests passing
- RC file management fully rebranded

**Phase 14 (CLI + Docs rebrand):** Blocked until 13-02 completes
- Shell wrappers now reference `ptt`, CLI must match
- Documentation will reference shell function as `ptt()`

## Self-Check: PASSED

Verified all claims:
- Created files: N/A (docs-only plans check skipped)
- Commits exist: 8ca91c0, ddb34bf ✓
- All files modified as documented ✓
- All tests passing ✓
- No old brand references remain in modified files ✓
