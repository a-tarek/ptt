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
- Location: `wt.zsh` lines 4-34 (`wt()` function)
- Contains: Main entry point with case statement
- Depends on: Individual command handler functions
- Used by: User invocation via shell

**Command Handlers:**
- Purpose: Implement specific worktree operations with full business logic
- Location: `wt.zsh` lines 36-364 (individual `_wt_*` functions)
- Contains: `_wt_new()`, `_wt_goto()`, `_wt_home()`, `_wt_init()`, `_wt_eject()`, `_wt_list()`, `_wt_merge()`, `_wt_rebase()`, `_wt_delete()`
- Key Functions:
  - `_wt_new()` (lines 36-86): Creates worktree with flag parsing for --copy/--symlink overrides
  - `_wt_goto()` (lines 88-103): Navigate to worktree by name
  - `_wt_home()` (lines 105-113): Return to main worktree
  - `_wt_init()` (lines 115-144): Scaffold .wtconfig template with commented examples
  - `_wt_eject()` (lines 146-276): Eject branch with flag parsing and stash management
  - `_wt_list()` (lines 278-310): Display all worktrees
  - `_wt_merge()` (lines 312-328): Merge worktree branch into current
  - `_wt_rebase()` (lines 330-346): Rebase current onto worktree branch
  - `_wt_delete()` (lines 348-364): Remove worktree
- Depends on: Helper functions, git CLI, filesystem operations
- Used by: Dispatcher router

**Helper Functions (Utilities):**
- Purpose: Provide common operations for worktree setup and path resolution
- Location: `wt.zsh` lines 366-505 (helper functions)
- Contains: Configuration processing, worktree path/branch resolution
- Key Functions:
  - `_wt_setup()` (lines 368-444): Apply .wtconfig actions with override mechanism using associative arrays
  - `_wt_resolve_path()` (lines 446-460): Find worktree filesystem path by name
  - `_wt_resolve_branch()` (lines 462-480): Find git branch by worktree name
  - `_wt_list_names()` (lines 482-505): Extract worktree short names for completion
- Depends on: Git porcelain output parsing, filesystem operations
- Used by: Command handlers

**Shell Completion System:**
- Purpose: Provide zsh autocompletion for commands and worktree arguments
- Location: `wt.zsh` lines 507-550 (`_wt()` completion function, `compdef` binding)
- Contains: Subcommand completion, dynamic worktree name completion, flag definitions (--copy/--symlink for new and eject)
- Depends on: Helper function `_wt_list_names()`, zsh completion system
- Used by: Interactive zsh shell

## Data Flow

**Worktree Creation Flow (`wt new`):**

1. User: `wt new [--copy <path>] [--symlink <path>] <name> [branch]`
2. Dispatcher: Routes to `_wt_new()` (line 36)
3. Flag parsing: Build overrides array in "action:path" format via `while [[ "$1" == --* ]]` loop
   - `--copy <path>` → `overrides+=("copy:<path>")`
   - `--symlink <path>` → `overrides+=("symlink:<path>")`
4. Validation: Git repository exists, target path doesn't exist
5. Git operation: `git worktree add <target_abs> -b <branch>` (create with new branch)
6. Fallback: If branch exists, retry `git worktree add <target_abs> <branch>`
7. Setup: Call `_wt_setup "$src_root" "$target_abs" "${overrides[@]}"`
   - Read `.wtconfig` file (if exists)
   - Apply actions (copy/symlink) with override precedence
   - Handle one-off overrides not in config
8. Terminal: `cd` into new worktree, display path and branch
9. Return: Success with confirmation messages

**Configuration Processing Flow (`_wt_setup`):**

1. Input: Source root, target worktree, override args in "action:path" format
2. Build overrides associative array: Parse each override arg, split on colon
3. Read `.wtconfig`: Line-by-line parsing (skip comments/blank lines)
4. For each config entry:
   - Parse action (copy/symlink) and path
   - Check if path has override → use override action instead
   - Mark override as applied
   - Execute action (cp -r or ln -s)
5. Apply one-off overrides: Process overrides not in config
6. Return: Setup complete with console output for each action

**Worktree Navigation Flow (`wt goto`):**

1. User: `wt goto <name>`
2. Dispatcher: Routes to `_wt_goto()`
3. Resolution: `_wt_resolve_path()` queries git worktree list, matches directory suffix
4. Navigation: `cd` into resolved path
5. Return: Success (shell changes directory)

**Branch Ejection Flow (`wt eject`):**

1. User: `wt eject [--copy <path>] [--symlink <path>] [name]` (from worktree with uncommitted changes)
2. Dispatcher: Routes to `_wt_eject()` (line 146)
3. Flag parsing: Build overrides array in "action:path" format via `while [[ "$1" == --* ]]` loop
4. Validation: Confirm not on detached HEAD, determine fallback branch (main/master or suffix-match)
5. Stash: Stash uncommitted changes (including untracked) before branch switch
6. Switch: `git checkout <fallback-branch>` in current worktree
7. Create: `git worktree add <new-path> <current-branch>` for ejected branch
8. Restore: Pop stash into new worktree via `git -C <new-path> stash pop`
9. Setup: Call `_wt_setup "$src_root" "$target_abs" "${overrides[@]}"`
   - Apply .wtconfig actions with override precedence
10. Navigate: `cd` into new worktree
11. Return: Success with branch confirmation

**State Management:**
- Git state: Maintained by `git worktree list --porcelain` queries
- Configuration: `.wtconfig` file at repository root (created via `wt init`)
- Config format: `<action> <path>` per line (copy or symlink)
- Override mechanism: Command-line flags (--copy/--symlink) take precedence over .wtconfig defaults
- Branch tracking: Git tracks branch association in worktree metadata

## Key Abstractions

**Worktree Identifier:**
- Purpose: Resolve user-provided worktree name to actual filesystem path and git branch
- Examples: `_wt_resolve_path()` (line 446), `_wt_resolve_branch()` (line 462)
- Pattern: Suffix matching on directory basename (e.g., user says "staging", matches "repo-staging")

**Configuration System:**
- Purpose: Define per-repository file handling strategy for new worktrees
- Implementation: `.wtconfig` file with `<action> <path>` syntax (lines 368-444 in `_wt_setup()`)
- Actions: `copy` (duplicate file) or `symlink` (link to source)
- Override mechanism: Associative arrays track overrides from command-line flags
- Pattern: Read config → apply overrides → execute actions → handle one-offs

**Git Worktree Interface:**
- Purpose: Abstract git worktree commands behind checked operations
- Pattern: `git worktree add`, `git worktree list --porcelain`, `git worktree remove` with error handling
- Fallback: Handle branch-already-exists case by retrying without `-b` flag (line 72)

**Porcelain Parser:**
- Purpose: Parse `git worktree list --porcelain` output into path/branch pairs
- Examples: Lines 288-302 in `_wt_list()`, lines 494-504 in `_wt_list_names()`
- Pattern: Line-by-line parsing of `worktree` and `branch` prefixed lines

**Fallback Branch Selection:**
- Purpose: Determine safe branch to switch to when ejecting current branch
- Logic: Home worktree → main/master; non-home → directory suffix match (lines 176-206)
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
- Not in git repository: `git rev-parse --git-dir` check (lines 54, 117, 156)
- Missing required arguments: Explicit usage message (lines 49-52, 90-93, 314-316)
- Target already exists: Pre-check with `-d` test (line 65, 219)
- Config already exists: Pre-check for .wtconfig (line 123)
- Returns: Non-zero exit code, user-facing error message

**Git Operation Failures:**
- Worktree add failure: Retry without `-b` flag (line 72), error if both fail (lines 74-77)
- Branch checkout failure: Rollback by restoring stash (lines 240-246)
- Returns: Non-zero exit code, attempt recovery

**Recovery/Rollback:**
- Eject operation: If checkout fails after stash, restore stash and abort (lines 240-246)
- If worktree creation fails, switch back to original branch and restore stash (lines 252-259)
- Intent: Preserve user work even on partial failure

## Cross-Cutting Concerns

**Logging:** Console output via `echo` statements indicating operation progress and completion

**Validation:** Input sanitization via zsh parameter expansion, git state validation via git commands

**Authentication:** Relies on user's git configuration and SSH keys; no credentials handled by script

**State Consistency:** All state changes go through git and filesystem (no internal state), queries happen at invocation time
