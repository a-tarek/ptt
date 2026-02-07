# Phase 4: Configuration System - Context

**Gathered:** 2026-02-07
**Status:** Ready for planning

<domain>
## Phase Boundary

Parse .wtconfig and handle copy/symlink/run actions when creating new worktrees. Support named config variants and CLI flag overrides as an alternative config source. Binary works correctly without any config (config-free operation).

</domain>

<decisions>
## Implementation Decisions

### Config format
- Plain text format: `action path` per line, `#` comments, blank lines ignored
- Three actions: `copy`, `symlink`, `run` — all ported from v1.0
- No glob patterns — explicit paths only
- Files and directories both supported (recursive copy for directories)
- Parent directories created automatically when needed (like `mkdir -p`)
- `run` syntax: everything after `run ` is the command string, no quoting required
- `run` commands execute with the new worktree (target) as working directory
- `run` output streamed to user (stdout/stderr visible in real-time)
- Execution order: strict sequential as written in the file — no type grouping
- Config files live at repo root: `.wtconfig`, `.wtconfig-*`

### Config model (mutually exclusive sources)
- One config source per invocation — no merging, no overriding:
  - `wt new <branch>` — uses `.wtconfig` (default)
  - `wt new <branch> --config <name>` — uses named config file
  - `wt new <branch> --copy X --symlink Y --run Z` — inline flags only, no config file read
- `--config` resolution: bare name → `.wtconfig-{name}` at repo root; contains `/` → treated as exact path
- Tab completion for `--config` lists available `.wtconfig-*` files (implementation in Phase 6)
- Inline flags can mix all three types freely: `--copy`, `--symlink`, `--run`
- Inline flag execution order: sequential in flag order as given on command line
- Duplicate flag paths = error (e.g., `--copy .env --symlink .env` is rejected)
- `--config nonexistent` = error, not fallback to no-config

### Init command updates
- `wt init` creates `.wtconfig` with template showing all three actions (copy, symlink, run) with examples
- `wt init --name foo` creates `.wtconfig-foo` with the same template
- Keep generic template — no project-type detection

### Failure & rollback
- Strict by default: missing source file for copy/symlink = failure
- Any action failure (copy, symlink, or run) aborts and rolls back
- Rollback = delete entire new worktree (git worktree remove + directory cleanup)
- Upfront validation: parse entire config and check all referenced files exist before executing any actions
- All validation errors reported at once (not fail-on-first)
- Run failures show both stderr output and exit code
- If rollback itself fails: warn with manual cleanup instructions, exit with error

### Config-free operation
- No .wtconfig + no flags = create worktree silently with zero setup actions
- No hint to run `wt init` — config is purely optional
- Success output: stream each action as it runs ("Copied .env", "Symlinked node_modules", "Running npm install..."), then show worktree path

### Claude's Discretion
- Exact validation error message formatting
- Internal config parsing implementation (structs, interfaces)
- How to detect mutually exclusive flag groups (config vs inline)
- Rollback implementation details (order of cleanup operations)

</decisions>

<specifics>
## Specific Ideas

- The v1.0 override model (merge flags with config) was explicitly rejected — too complex with edge cases
- Config model inspired by simplicity: "pick one source" eliminates an entire class of bugs
- Strict-by-default catches typos and stale config entries that silent-skip would hide
- "What you write is what you get" — execution order matches file/flag order exactly

</specifics>

<deferred>
## Deferred Ideas

- **copyEnv action** — A config action that copies env files but prompts for variable overrides per worktree (e.g., `copyEnv .env.local --VITE_PORT`). Would enable different ports/credentials per worktree without manual editing. Significant feature with prompt handling, env file parsing, variable substitution — belongs in its own phase.

</deferred>

---

*Phase: 04-configuration-system*
*Context gathered: 2026-02-07*
