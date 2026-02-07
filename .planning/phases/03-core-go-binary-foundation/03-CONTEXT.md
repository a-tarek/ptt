# Phase 3: Core Go Binary Foundation - Context

**Gathered:** 2026-02-07
**Status:** Ready for planning

<domain>
## Phase Boundary

Establish the Go project structure with build system and implement simple commands that don't require shell integration: `wt --version`, `wt list`, `wt init`, and `wt delete`. Proper exit codes and error handling. Shell wrappers and cd-based commands are separate phases.

</domain>

<decisions>
## Implementation Decisions

### List output format
- Default columns: name, branch, dirty/clean status indicator
- Path shown only with a flag (e.g., `-a` or `--all`)
- Current worktree marked with asterisk prefix (`*`) — like `git branch`
- Color auto-detected: color when stdout is a terminal, plain when piped
- Empty state: silent (no output, exit 0) — script-friendly

### Delete behavior
- No confirmation for clean worktrees — deletes immediately
- Dirty worktrees prompt for confirmation: "Worktree 'foo' has uncommitted changes. Delete? [y/N]"
- `--force` skips all confirmation, even for dirty worktrees
- Branch NOT deleted by default — `--branch` flag to also delete the branch
- Cannot delete current worktree — error: "can't delete current worktree"
- Single worktree per invocation (no multi-delete)
- Silent on success — no output on successful delete

### Error & output style
- Terse and direct — git-style messages
- Errors prefixed with `error:` (e.g., `error: worktree 'foo' not found`)
- Errors to stderr, data to stdout
- Simple commands silent on success (init, delete)
- Multi-step commands report each action (copy, symlink, run) — applies to Phase 5's `wt new`
- No `--quiet` flag — commands are already minimal

### Init defaults
- Template contains commented-out examples showing copy/symlink/run syntax
- Config format uses abstract actions: `copy`, `symlink` (Go handles cross-platform), plus `run` as escape hatch for custom commands
- Error if .wtconfig already exists — no overwrite (no --force)
- Created in current directory (not necessarily repo root)
- Requires being inside a git repo — error if not
- No auto-detection of project files — just the template

### Claude's Discretion
- Dirty/clean status indicator style (symbol choice)
- Exact list column alignment and spacing
- Commented example content in .wtconfig template
- Go project structure (module layout, package organization)
- Build system and CI tooling choices

</decisions>

<specifics>
## Specific Ideas

- Config format should use abstract actions for portability: `copy .env`, `symlink node_modules`, `run npm install` — not raw shell commands
- The `run` action is an escape hatch for anything copy/symlink doesn't cover
- User wants to preserve the action-reporting output style from current wt.zsh (e.g., "copy .env.local", "symlink node_modules") for multi-step commands in later phases

</specifics>

<deferred>
## Deferred Ideas

- Config format details and parsing logic — Phase 4 (Configuration System)
- Multi-step action reporting for `wt new` — Phase 5 (Directory-Changing Commands)
- Cross-platform shell commands via `run` action — Phase 4/5

</deferred>

---

*Phase: 03-core-go-binary-foundation*
*Context gathered: 2026-02-07*
