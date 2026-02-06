# Coding Conventions

**Analysis Date:** 2026-02-07

## Naming Patterns

**Functions:**
- Private helper functions use underscore prefix: `_wt_new`, `_wt_goto`, `_wt_resolve_path`
- Public entry point uses simple name: `wt`
- Function names use snake_case throughout
- Helper functions group under their subsystem: `_wt_*` functions are all worktree-related helpers

**Variables:**
- Local variables use snake_case: `copy_nm`, `src_root`, `target_abs`, `current_branch`
- Boolean flags use descriptive names: `copy_nm` (copy node modules), `did_stash`
- Path variables use descriptive suffixes: `_abs` for absolute paths, `_path` for directory paths
- Temporary variables explicitly named: `base_name`, `repo_basename`, `stash_before`, `stash_after`

**Constants:**
- Error messages use clear, actionable descriptions: `"Error: not inside a git repository"`
- Command descriptions in completion use sentence format: `'new:Create a new worktree'`
- Branch names are lowercase: `main`, `master`

## Code Style

**Shell-specific:**
- Zsh functions use `local` keyword for all variables to prevent global scope pollution
- Code uses `[[ ]]` for conditional tests (Zsh-specific, more robust than `[ ]`)
- Variables are quoted to handle spaces: `"$src_root"`, `"$target_abs"`
- Command substitution uses `$()` syntax instead of backticks: `$(git rev-parse --show-toplevel)`

**Indentation:**
- Consistent 2-space indentation throughout file
- Case statements use standard shell conventions with conditions at same indent level

**Whitespace:**
- Single blank line between function definitions
- Comments use `#` with space: `# Comment` not `#Comment`
- Multi-line strings and heredocs not used (kept simple)

## Error Handling

**Patterns:**
- Explicit validation before operations: Check git directory existence before attempting git operations
- Early returns on validation failure: `if [[ -z "$name" ]]; then echo "..."; return 1; fi`
- Conditional execution with `||`: `git worktree add "$target_abs" -b "$branch" 2>/dev/null || fallback_approach`
- Explicit error states checked: `if (( stash_after > stash_before ))`
- Rollback on partial failure in `_wt_eject`: stash state tracked and restored on checkout failure

**Examples from codebase:**
- `_wt_new`: Validates git repository, checks target doesn't exist, handles git command failure with fallback
- `_wt_eject`: Complex multi-step process with state tracking for stash management and rollback on failure
- `_wt_resolve_path`: Silently returns empty on failure, letting caller handle error message

## Logging & User Feedback

**Pattern:**
- Status messages use `echo` for informational output
- Error messages consistently prefixed with `"Error: "` for distinction
- Progress messages indicate action being taken: `"Creating worktree: ..."`, `"Copying node_modules..."`
- Summary output after major operations: `"Ready: $(pwd)"`, `"Branch: $(git branch --show-current)"`
- Errors sent to stderr implicitly through `echo` (not redirected)

**Examples:**
- Information: `echo "Creating worktree: ${repo_basename}-${name} (branch: $branch)"`
- Error: `echo "Error: $target_abs already exists"`
- Summary: Two-line output with path and branch confirmation

## Git Integration

**Conventions:**
- All git commands use `2>/dev/null` to suppress noise, except when error visibility is needed
- Git porcelain format used for parsing: `git worktree list --porcelain`
- Branch references use full paths: `refs/heads/main` not just `main`
- Git directory checks use: `if ! git rev-parse --git-dir &>/dev/null`

## Function Design

**Size:** Functions range 10-100 lines, with most helper functions under 40 lines
- `_wt_new`: 67 lines (multi-step creation)
- `_wt_eject`: 129 lines (complex state management)
- `_wt_resolve_path`: 13 lines (simple lookup)

**Parameters:**
- Functions typically take 1-2 required parameters
- Optional parameters passed as varargs: `"$@"`
- Flag parsing done explicitly in calling function

**Return Values:**
- Exit status used for success/failure: `return 0` or `return 1`
- Data returned via stdout for parsing: `echo "$wt_entry"` in resolve functions
- Complex operations set local variables for same-scope return

## Comments

**When Used:**
- Section headers for logical groupings: `# --- Helpers ---`, `# --- Zsh Completions ---`
- Inline explanations for non-obvious logic: `# Strip repo basename prefix to get the suffix`
- State transitions in complex flows: `# 1. Get current branch`, `# 5. Stash uncommitted changes`

**Style:**
- Single-line comments only (no multi-line blocks)
- Comments placed above the lines they describe
- Numbered steps in multi-step operations for clarity

## Module Organization

**Dispatch pattern:**
- Main entry point (`wt`) uses case statement for command routing
- Each subcommand maps to dedicated function: `new` → `_wt_new`
- Shared helpers grouped at bottom under `# --- Helpers ---`
- Completion logic separated under `# --- Zsh Completions ---`

**Exports:**
- Main function `wt` is exported/available globally
- Helper functions with `_` prefix are internal (not meant for direct use)
- Completion function `_wt` is registered with `compdef _wt wt`

---

*Convention analysis: 2026-02-07*
