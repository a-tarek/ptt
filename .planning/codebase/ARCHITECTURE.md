# Architecture

**Analysis Date:** 2026-02-07

## Pattern Overview

**Overall:** Command Dispatcher Pattern with Modular Helper Functions

**Key Characteristics:**
- Single-file shell script architecture
- Command router dispatching to focused subcommand handlers
- Functional decomposition with helper utilities
- Pure shell implementation (zsh) with minimal external dependencies
- Stateless command execution relying on git state

## Layers

**Command Dispatcher (Router):**
- Purpose: Parse user input and route to appropriate command handler
- Location: `wt.zsh` lines 4-31 (`wt()` function)
- Contains: Main entry point with case statement
- Depends on: Individual command handler functions
- Used by: User invocation via shell

**Command Handlers:**
- Purpose: Implement specific worktree operations with full business logic
- Location: `wt.zsh` lines 34-346 (individual `_wt_*` functions)
- Contains: `_wt_new()`, `_wt_goto()`, `_wt_home()`, `_wt_eject()`, `_wt_list()`, `_wt_merge()`, `_wt_rebase()`, `_wt_delete()`
- Depends on: Helper functions, git CLI, filesystem operations
- Used by: Dispatcher router

**Helper Functions (Utilities):**
- Purpose: Provide common path resolution and state lookup operations
- Location: `wt.zsh` lines 348-409 (`_wt_resolve_path()`, `_wt_resolve_branch()`, `_wt_list_names()`)
- Contains: Worktree path/branch resolution, name parsing logic
- Depends on: Git porcelain output parsing
- Used by: Command handlers

**Shell Completion System:**
- Purpose: Provide zsh autocompletion for commands and worktree arguments
- Location: `wt.zsh` lines 411-446 (`_wt()` completion function, `compdef` binding)
- Contains: Subcommand completion, dynamic worktree name completion, flag definitions
- Depends on: Helper function `_wt_list_names()`, zsh completion system
- Used by: Interactive zsh shell

## Data Flow

**Worktree Creation Flow (`wt new`):**

1. User: `wt new [--copy-node-modules] <name> [branch]`
2. Dispatcher: Routes to `_wt_new()`
3. Handler validates: Git repository exists, target path doesn't exist
4. Git operation: `git worktree add <path> -b <branch>` (create with new branch)
5. Fallback: If branch exists, retry `git worktree add <path> <branch>`
6. State management: Copy `.env.local` if present, symlink (or copy) `node_modules`
7. Terminal: `cd` into new worktree, display path and branch
8. Return: Success with confirmation messages

**Worktree Navigation Flow (`wt goto`):**

1. User: `wt goto <name>`
2. Dispatcher: Routes to `_wt_goto()`
3. Resolution: `_wt_resolve_path()` queries git worktree list, matches directory suffix
4. Navigation: `cd` into resolved path
5. Return: Success (shell changes directory)

**Branch Ejection Flow (`wt eject`):**

1. User: `wt eject [name]` (from worktree with uncommitted changes)
2. Dispatcher: Routes to `_wt_eject()`
3. Validation: Confirm not on detached HEAD, determine fallback branch (main/master or suffix-match)
4. Stash: Stash uncommitted changes before branch switch
5. Switch: `git checkout <fallback-branch>` in current worktree
6. Create: `git worktree add <new-path> <current-branch>` for ejected branch
7. Restore: Pop stash into new worktree via `git -C <new-path> stash pop`
8. Setup: Copy `.env.local` and symlink `node_modules` in new worktree
9. Navigate: `cd` into new worktree
10. Return: Success with branch confirmation

**State Management:**
- Git state: Maintained by `git worktree list --porcelain` queries
- Environment: `.env.local` copied to each worktree (git-ignore safe)
- Dependencies: `node_modules` symlinked (or copied with `--copy-node-modules`)
- Branch tracking: Git tracks branch association in worktree metadata

## Key Abstractions

**Worktree Identifier:**
- Purpose: Resolve user-provided worktree name to actual filesystem path and git branch
- Examples: `_wt_resolve_path()` (line 352), `_wt_resolve_branch()` (line 367)
- Pattern: Suffix matching on directory basename (e.g., user says "staging", matches "repo-staging")

**Git Worktree Interface:**
- Purpose: Abstract git worktree commands behind checked operations
- Pattern: `git worktree add`, `git worktree list --porcelain`, `git worktree remove` with error handling
- Fallback: Handle branch-already-exists case by retrying without `-b` flag (line 72)

**Porcelain Parser:**
- Purpose: Parse `git worktree list --porcelain` output into path/branch pairs
- Examples: Lines 270-291 in `_wt_list()`, lines 398-408 in `_wt_list_names()`
- Pattern: Line-by-line parsing of `worktree` and `branch` prefixed lines

**Fallback Branch Selection:**
- Purpose: Determine safe branch to switch to when ejecting current branch
- Logic: Home worktree → main/master; non-home → directory suffix match (lines 149-181)
- Safety: Validates branch exists before attempting switch

## Entry Points

**User Command (`wt`):**
- Location: `wt.zsh` line 4
- Triggers: User invokes `wt <command> [args]` in shell
- Responsibilities: Parse command name, route to handler, handle unknown commands

**Completion Entry (`_wt`):**
- Location: `wt.zsh` line 413
- Triggers: Zsh completion system on tab completion
- Responsibilities: Provide context-aware command and argument suggestions

## Error Handling

**Strategy:** Fail-safe with informative error messages and optional rollback

**Patterns:**

**Validation Errors:**
- Not in git repository: `git rev-parse --git-dir` check (lines 52, 130)
- Missing required arguments: Explicit usage message (lines 48-50, 105-106)
- Target already exists: Pre-check with `-d` test (line 63)
- Returns: Non-zero exit code, user-facing error message

**Git Operation Failures:**
- Worktree add failure: Retry without `-b` flag (line 72), error if both fail (line 73)
- Branch checkout failure: Rollback by restoring stash (lines 216-220)
- Returns: Non-zero exit code, attempt recovery

**Recovery/Rollback:**
- Eject operation: If checkout fails after stash, restore stash and abort (lines 214-220)
- If worktree creation fails, switch back to original branch and restore stash (lines 227-232)
- Intent: Preserve user work even on partial failure

## Cross-Cutting Concerns

**Logging:** Console output via `echo` statements indicating operation progress and completion

**Validation:** Input sanitization via zsh parameter expansion, git state validation via git commands

**Authentication:** Relies on user's git configuration and SSH keys; no credentials handled by script

**State Consistency:** All state changes go through git and filesystem (no internal state), queries happen at invocation time
