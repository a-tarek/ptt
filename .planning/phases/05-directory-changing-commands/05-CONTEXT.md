# Phase 5: Directory-Changing Commands - Context

**Gathered:** 2026-02-07
**Status:** Ready for planning

<domain>
## Phase Boundary

Implement all commands that change directories or create/manage worktrees: `wt new`, `wt goto`, `wt home`, `wt eject`, `wt merge`, `wt rebase`. The Go binary outputs structured data (path on stdout, messages on stderr) for shell wrapper coordination. Shell wrappers themselves are Phase 6.

</domain>

<decisions>
## Implementation Decisions

### Shell coordination protocol
- `--output-path` hidden flag: when passed, binary outputs ONLY the target path to stdout (machine-readable)
- Without the flag: binary outputs human-friendly messages to stdout (direct invocation mode)
- Shell wrapper always passes `--output-path`; users never type it manually
- Confirmation messages (e.g., "Switched to feature-x (branch: feature-x, clean)") printed to stderr by the binary — includes worktree name, branch, and dirty status
- Errors go to stderr with non-zero exit code; wrapper stays in current directory
- `--output-path` flag is hidden from `--help` and completions
- Already-there case: no-op with "Already in feature-x" message to stderr, exit 0

### Worktree placement (auto-detect mode)
- **Bare repo** → nested mode: worktrees created under the bare repo root (e.g., `/code/wt/staging`)
- **Regular repo** → sibling mode: worktrees created as siblings (e.g., `/code/wt-staging`)
- Sibling mode fixes v1.0 compounding bug: always resolves to the original repo name, so `wt new feat-1-a` from inside `wt-feat-1` creates `/code/wt-feat-1-a` (not `/code/wt-feat-1-feat-1-a`)
- Auto-detected via `git rev-parse --is-bare-repository` — no config flag needed

### wt goto
- Argument required — errors with "worktree name required" if no argument (tab completion in Phase 6 makes discovery easy)
- Uses existing suffix-match resolution from Phase 3
- Worktrees-only context — errors with "not in a worktree setup" if not in a bare/worktree repo

### wt home
- Dedicated command (not an alias for `wt goto main`)
- Navigates to the bare repo root directory
- Same error handling as other cd commands (worktrees-only context)

### wt new
- Accepts branch name as required argument: `wt new feature-login`
- Always branches from HEAD (no `--base` flag)
- Local branches only — no auto-detection of remote branches, no auto-fetch. User must fetch/create local branches themselves
- If branch doesn't exist locally, creates a new branch from HEAD
- If branch exists locally, uses it (tries `-b` first, falls back to existing)
- Auto-cd into the new worktree after creation
- Applies `.wtconfig` by default if present, silently skips if absent
- Rollback everything on config action failure (remove worktree entirely)
- Detailed confirmation message: "Created worktree wt-staging (branch: staging)\nApplied .wtconfig (3 actions)\nSwitched to wt-staging (branch: staging, clean)"

### wt new flags
- `--config <name>`: use `.wtconfig-<name>` instead of `.wtconfig` (e.g., `--config ci` → `.wtconfig-ci`)
- `--skip-config`: skip all config actions even if `.wtconfig` exists
- `--copy`/`--symlink`: inline overrides (merge with whatever config is loaded)
- Flag hierarchy: default `.wtconfig` < `--config` override < `--skip-config` bypass; `--copy`/`--symlink` merge on top

### wt eject
- Replicate v1.0 flow exactly: stash (including untracked) → switch to fallback branch → create new worktree → pop stash → apply config → cd into new worktree
- Fallback branch detection: home worktree → main/master; non-home worktree → branch matching directory suffix
- Improvement over v1.0: warn when stash pop causes merge conflicts ("Stash restored with conflicts — resolve before committing")
- Supports same flags as wt new: `--config`, `--skip-config`, `--copy`, `--symlink`
- Rollback on failure at each step (same as v1.0)

### wt merge / wt rebase
- Simple wrappers: resolve worktree name to branch, run `git merge`/`git rebase`
- No special conflict handling — let git handle it naturally
- Pure binary commands — no shell wrapper coordination needed (no cd involved)
- Convenience is worktree name → branch name resolution

### Claude's Discretion
- Exact error message wording
- Internal code organization (how commands share resolution logic)
- Test structure and coverage strategy
- Whether merge/rebase need the `--output-path` flag (likely not since no cd)

</decisions>

<specifics>
## Specific Ideas

- Confirmation after cd should include branch name and dirty status: "Switched to feature-x (branch: feature-x, clean)"
- v1.0 eject behavior is well-designed — port it faithfully with the stash conflict warning improvement
- The auto-detect bare vs regular repo for worktree placement is a v2.0 improvement over v1.0

</specifics>

<deferred>
## Deferred Ideas

- `wt graph` — show worktree dependencies in a visual graph (new capability, own phase)

</deferred>

---

*Phase: 05-directory-changing-commands*
*Context gathered: 2026-02-07*
