# ptt

## What This Is

A cross-platform Git worktree manager ("a potato worktree manager"). Go binary with multi-shell support (bash/zsh/fish) across Linux, macOS, and Windows WSL. Distributed via npm as `@a-tarek/ptt` with platform-specific binaries.

## Core Value

A single `ptt` command that works in any shell on any platform with full autocompletion — managing git worktrees should be effortless everywhere.

## Current Milestone: v2.0 Pre-Release (Bare Repo + Command Polish)

**Goal:** Add bare repo conversion and nested worktree support, rename go→cd, ship v2.0.

**Target changes:**
- `ptt mk-bare-repo`: convert normal clone to bare repo with nested worktrees (creates `-bare` copy)
- `ptt mk` in bare repos: create worktrees nested inside bare repo directory
- Rename `go` → `cd` (`ptt cd <wt>` navigates, `ptt cd` goes to main worktree), `go` kept as alias
- `.pttconfig/` lives at bare repo root, shared by all worktrees

## Requirements

### Validated

- ✓ Internal codebase map (7 files) — v1.0 docs milestone
- ✓ README.md with full user-facing documentation — v1.0 docs milestone
- ✓ Go binary with all 9 commands ported — v2.0 rewrite
- ✓ Rebrand wt → ptt (binary, module, npm, shell wrappers) — v2.0 rebrand
- ✓ Command restructure: mk, go, rm, ls as primary names — v2.0 rebrand
- ✓ Config directory: .pttconfig/ with default and named configs — v2.0 rebrand
- ✓ Shell wrappers + npm distribution under @a-tarek/ptt — v2.0 rebrand

### Active

- [ ] `ptt mk-bare-repo`: convert normal clone to bare repo with nested worktrees
- [ ] `ptt mk` nests worktrees inside bare repo directory when in bare repo context
- [ ] Rename `go` → `cd` as primary command name (`go` kept as alias)
- [ ] `ptt cd` (no args) always navigates to main worktree (bare and non-bare)
- [ ] `.pttconfig/` at bare repo root, shared across all worktrees

### Out of Scope

- Claude Code skill integration — Claude can run git commands natively, minimal added value
- PowerShell support — focus on bash/zsh/fish for v2.0
- Plugin manager install docs — npm handles distribution now
- Backward compatibility migration tool — clean break before first release

## Context

- 10 commands (restructured): mk, mk-bare-repo, cd, eject, init, ls, merge, rebase, rm + install/uninstall/shell-init
- .pttconfig/ directory — default config + named configs via --config flag; lives at bare repo root in bare repos
- --copy/--symlink override flags on `mk` and `eject`
- Commands that change directory (cd, mk, eject) need thin shell wrapper functions — a subprocess cannot change the parent shell's directory
- `cd` command (renamed from `go`): with args navigates to worktree, without args navigates to main worktree
- Bare repo support: `mk-bare-repo` converts normal clone to bare structure, `mk` nests worktrees inside bare repos
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
| Merge goto+home into go | Simpler mental model: `ptt go` = home, `ptt go <wt>` = navigate | ⚠️ Revisit — renaming to `cd` |
| Config directory (.pttconfig/) | Named configs in directory vs flat files — cleaner, supports --config flag | — Pending |
| @a-tarek/ptt npm scope | Personal scope, guarantees npm availability | — Pending |
| github.com/a-tarek/ptt module | Matches npm scope, personal GitHub | — Pending |

| Rename go → cd | `cd` is more intuitive for directory navigation; `go` kept as alias | — Pending |
| Bare repo support | Nested worktrees inside bare repo avoids cluttering parent directory | — Pending |
| mk-bare-repo as copy | Safer than in-place restructure — user verifies then deletes old repo | — Pending |
| .pttconfig at bare root | Project-level config shared by all worktrees, not per-worktree | — Pending |

---
*Last updated: 2026-02-09 after bare repo + cd rename milestone planning*
