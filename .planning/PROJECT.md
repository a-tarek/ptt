# ptt

## What This Is

A cross-platform Git worktree manager ("a potato worktree manager"). Go binary with multi-shell support (bash/zsh/fish) across Linux, macOS, and Windows WSL. Distributed via npm as `@a-tarek/ptt` with platform-specific binaries.

## Core Value

A single `ptt` command that works in any shell on any platform with full autocompletion — managing git worktrees should be effortless everywhere.

## Current Milestone: v2.0 Pre-Release (Rebrand + Command Restructure)

**Goal:** Rebrand wt → ptt, restructure commands for better UX, and move config to directory structure before first public release.

**Target changes:**
- Rename binary/module/npm from `wt` to `ptt` (`github.com/a-tarek/ptt`, `@a-tarek/ptt`)
- Command restructure: new→mk, goto+home→go, delete→rm, list→ls (with backward compat aliases)
- Config directory: `.wtconfig*` flat files → `.pttconfig/` directory with named configs
- Shell wrappers/installer updated for ptt branding

## Requirements

### Validated

- ✓ Internal codebase map (7 files) — v1.0 docs milestone
- ✓ README.md with full user-facing documentation — v1.0 docs milestone

### Active

- [ ] Rename binary/module/npm: wt → ptt (github.com/a-tarek/ptt, @a-tarek/ptt)
- [ ] Command restructure: mk, go (merged goto+home), rm, ls as primary names
- [ ] Config directory: .pttconfig/ with default and named configs (--config flag)
- [ ] Shell wrappers and installer updated for ptt branding
- [ ] npm distribution updated for @a-tarek scope

### Out of Scope

- Claude Code skill integration — Claude can run git commands natively, minimal added value
- PowerShell support — focus on bash/zsh/fish for v2.0
- Plugin manager install docs — npm handles distribution now
- Backward compatibility migration tool — clean break before first release

## Context

- 9 commands (restructured): mk, go, eject, init, ls, merge, rebase, rm + install/uninstall/shell-init
- .pttconfig/ directory replaces flat .wtconfig files — default config + named configs via --config flag
- --copy/--symlink override flags on `mk` and `eject`
- Commands that change directory (go, mk, eject) need thin shell wrapper functions — a subprocess cannot change the parent shell's directory
- `go` command merges old goto+home: with args navigates to worktree, without args navigates home
- Standard pattern used by zoxide, nvm, direnv: executable does logic, sourced shell function handles cd
- npm: @a-tarek/ptt with platform packages @a-tarek/ptt-{platform}-{arch}
- cobra CLI framework provides built-in completion generation for bash/zsh/fish
- esbuild/turbo/biome precedent for distributing Go binaries via npm
- Go module: github.com/a-tarek/ptt

## Constraints

- **Language**: Go — single binary, fast startup (~5ms), cobra completions built-in
- **Shell wrapper compat**: Must work in bash 3.2 (macOS default), zsh, fish
- **Distribution**: npm scoped package (@a-tarek/ptt) — binary download on postinstall
- **Backward compat**: Old command names (new, goto, delete, list) kept as aliases
- **Pre-release**: No public users yet — clean break is safe

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Source-only install (v1.0) | Simplest approach for zsh-only tool | ✓ Good |
| Container tips in README (v1.0) | Practical value for Docker users | ✓ Good |
| Go over Node.js/bash | Fast startup (~5ms vs ~150ms), single binary, cobra completions for free | — Pending |
| npm for distribution | Download tracking, familiar install (npx), cross-platform binary delivery | — Pending |
| Shell wrappers for cd | Subprocess can't change parent directory — standard pattern (zoxide, nvm) | — Pending |
| Scoped npm package | "wt" taken on npm; @scope/wt guarantees availability, CLI stays `wt` | ⚠️ Revisit — rebranding to @a-tarek/ptt |
| Drop Claude skill | Claude runs git commands natively, skill adds minimal value | — Pending |
| Keep wt.zsh as legacy | Existing users shouldn't break, gradual migration | — Pending |
| Port only, no new features | Clean port reduces risk, new features come after v2.0 | — Pending |
| Target bash 3.2 for wrappers | Maximum macOS compatibility without requiring brew install bash | — Pending |
| Rebrand wt → ptt | "a potato worktree manager" — distinctive name, clean break before first release | — Pending |
| Merge goto+home into go | Simpler mental model: `ptt go` = home, `ptt go <wt>` = navigate | — Pending |
| Config directory (.pttconfig/) | Named configs in directory vs flat files — cleaner, supports --config flag | — Pending |
| @a-tarek/ptt npm scope | Personal scope, guarantees npm availability | — Pending |
| github.com/a-tarek/ptt module | Matches npm scope, personal GitHub | — Pending |

---
*Last updated: 2026-02-08 after v2.0 rebrand planning*
