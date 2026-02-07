# Phase 6: Shell Integration - Context

**Gathered:** 2026-02-07
**Status:** Ready for planning

<domain>
## Phase Boundary

Provide shell wrapper functions and tab completions for bash, zsh, and fish that coordinate with the Go binary's `--output-path` protocol to change directories in the user's current shell. The wrapper intercepts `wt` commands, delegates to the binary, and performs `cd` when needed. Completions are generated via Cobra's built-in completion system.

npm distribution (Phase 7) and the interactive installer (Phase 8) are separate phases.

</domain>

<decisions>
## Implementation Decisions

### Wrapper sourcing & setup
- `wt shell-init` command outputs wrapper code for the detected shell
- Auto-detects current shell (no explicit argument needed)
- User adds `eval $(wt shell-init)` to their rc file (or installer does it in Phase 8)
- Completions are a separate command: `wt completion <shell>` (two eval lines, not bundled)
- No customization flags (no `--cmd` alias, no `--no-alias`)
- Function is always named `wt`
- No legacy wt.zsh detection — shell-init is standalone, ignores existing wt.zsh
- No env vars set after cd

### cd command behavior
- On successful cd: show worktree info (name + branch + path) then change directory
- Errors follow existing Go binary error patterns (consistent with Phase 3-5)
- Current stdout/stderr protocol is kept: binary prints path to stdout via `--output-path`, info/errors to stderr, wrapper captures stdout for cd
- `wt new` shows setup summary after creation (config actions that ran + worktree info)
- `wt eject` auto-cd's into the new worktree after completion
- "Already there" case: just show message, no redundant cd
- `wt merge`/`wt rebase`: let git output pass through naturally, no wrapper formatting
- No environment variables set after cd

### Tab completion
- Use Cobra's built-in completion generation for all shells
- Dynamic completions: every tab press queries git worktree list (always accurate, no caching)
- Worktree name completions show names only (no branch descriptions)
- `wt new` has no completions for the branch name argument (it's a new name)

### Multi-shell parity
- All three shells (bash, zsh, fish) supported in this phase — no deferral
- Identical user-facing behavior across all shells — no shell-specific differences
- No shell mismatch validation in shell-init

### Claude's Discretion
- All wt commands route through wrapper vs only cd commands (leaning: all through wrapper — simpler mental model, like zoxide)
- Wrapper calls binary via `command wt` to bypass function (standard pattern)
- Wrapper scripts embedded in Go binary via `go:embed` (leaning: embed — single binary distribution, no missing files)
- Minimum bash version support (leaning: bash 3.2+ for macOS compatibility, wrapper is simple enough)
- Exact worktree info format on successful cd

</decisions>

<specifics>
## Specific Ideas

- `eval $(wt shell-init)` pattern follows zoxide, direnv, starship conventions — familiar to users
- Completions and wrapper are separate eval lines so users can opt into completions independently
- The installer in Phase 8 will handle adding these lines to rc files automatically

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 06-shell-integration*
*Context gathered: 2026-02-07*
