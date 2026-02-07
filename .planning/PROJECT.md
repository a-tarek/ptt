# wt

## What This Is

A cross-platform Git worktree manager. Currently a zsh script (`wt.zsh`, ~550 lines), being rewritten in Go for multi-shell support (bash/zsh/fish) across Linux, macOS, and Windows WSL. Distributed via npm as a scoped package with download tracking.

## Core Value

A single `wt` command that works in any shell on any platform with full autocompletion — managing git worktrees should be effortless everywhere.

## Current Milestone: v2.0 Go Rewrite

**Goal:** Rewrite wt in Go for cross-platform, multi-shell support with npm distribution.

**Target features:**
- Go binary porting all 9 current commands
- Shell wrappers for bash/zsh/fish (cd handling)
- cobra-generated completions for all 3 shells
- npm scoped package distribution (@scope/wt)
- Interactive installer (`npx @scope/wt install`)

## Requirements

### Validated

- ✓ Internal codebase map (7 files) — v1.0 docs milestone
- ✓ README.md with full user-facing documentation — v1.0 docs milestone

### Active

- [ ] Go binary porting all wt.zsh commands (new, goto, home, init, eject, list, merge, rebase, delete)
- [ ] Shell wrappers for cd-handling in bash, zsh, fish
- [ ] cobra-generated completions for bash, zsh, fish
- [ ] npm scoped package with platform-specific binary distribution
- [ ] Interactive installer (`npx @scope/wt install`) — shell detection, completions setup
- [ ] Download tracking via npm
- [ ] Legacy wt.zsh preserved alongside

### Out of Scope

- Claude Code skill integration — Claude can run git commands natively, minimal added value
- PowerShell support — focus on bash/zsh/fish for v2.0
- New features beyond current wt.zsh — port only, no additions
- Plugin manager install docs — npm handles distribution now

## Context

- 9 commands to port: new, goto, home, init, eject, list, merge, rebase, delete
- .wtconfig support for per-repo file handling strategy (copy/symlink actions)
- --copy/--symlink override flags on `new` and `eject`
- Commands that change directory (goto, home, new, eject) need thin shell wrapper functions — a subprocess cannot change the parent shell's directory
- Standard pattern used by zoxide, nvm, direnv: executable does logic, sourced shell function handles cd
- npm package name "wt" is taken; scoped @scope/wt guarantees availability, CLI command stays `wt`
- cobra CLI framework provides built-in completion generation for bash/zsh/fish
- esbuild/turbo/biome precedent for distributing Go binaries via npm

## Constraints

- **Language**: Go — single binary, fast startup (~5ms), cobra completions built-in
- **Shell wrapper compat**: Must work in bash 3.2 (macOS default), zsh, fish
- **Distribution**: npm scoped package — binary download on postinstall
- **Behavior**: Exact feature parity with wt.zsh — no new features, no removed features
- **Legacy**: wt.zsh kept alongside for existing zsh-only users

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Source-only install (v1.0) | Simplest approach for zsh-only tool | ✓ Good |
| Container tips in README (v1.0) | Practical value for Docker users | ✓ Good |
| Go over Node.js/bash | Fast startup (~5ms vs ~150ms), single binary, cobra completions for free | — Pending |
| npm for distribution | Download tracking, familiar install (npx), cross-platform binary delivery | — Pending |
| Shell wrappers for cd | Subprocess can't change parent directory — standard pattern (zoxide, nvm) | — Pending |
| Scoped npm package | "wt" taken on npm; @scope/wt guarantees availability, CLI stays `wt` | — Pending |
| Drop Claude skill | Claude runs git commands natively, skill adds minimal value | — Pending |
| Keep wt.zsh as legacy | Existing users shouldn't break, gradual migration | — Pending |
| Port only, no new features | Clean port reduces risk, new features come after v2.0 | — Pending |
| Target bash 3.2 for wrappers | Maximum macOS compatibility without requiring brew install bash | — Pending |

---
*Last updated: 2026-02-07 after v2.0 milestone initialization*
