# wt — Git Worktree Manager

A zsh tool that manages git worktrees with shared configuration, file copying/symlinking, and branch ejection.

**Features:**
- Create worktrees with shared configuration (via `.wtconfig`)
- Copy or symlink files between worktrees automatically
- Eject branches into dedicated worktrees
- Navigate between worktrees seamlessly
- Merge and rebase across worktrees

## Installation

```zsh
# Clone the repository
git clone https://github.com/aatarek/wt.git ~/wt

# Add to your .zshrc
echo 'source ~/wt/wt.zsh' >> ~/.zshrc

# Reload shell
source ~/.zshrc
```

**Requirements:** zsh and git 2.5+ (with worktree support).

## Quick Start

```zsh
wt init                          # Create .wtconfig in your repo
wt new feature-login             # Create worktree + branch
wt goto feature-login            # Jump to a worktree
wt home                          # Return to main worktree
wt delete feature-login          # Clean up when done
```

**What each command does:**
- `wt init` — Creates a `.wtconfig` template to define which files should be copied or symlinked into new worktrees
- `wt new feature-login` — Creates a new worktree named `<repo>-feature-login` with a branch `feature-login`, then applies `.wtconfig` rules
- `wt goto feature-login` — Changes directory into the `feature-login` worktree
- `wt home` — Changes directory back to the main worktree
- `wt delete feature-login` — Removes the worktree (keeps the branch)

## Commands

```
Usage: wt <command> [args]

Commands:
  new [flags] <name> [branch]                  Create a new worktree
  goto <worktree>                              cd into a worktree
  home                                         cd into the main worktree
  init                                         Create .wtconfig template
  eject [flags] [name]                         Eject current branch into its own worktree
  list                                         List all worktrees
  merge <worktree>                             Merge worktree's branch into current
  rebase <worktree>                            Rebase current onto worktree's branch
  delete <worktree>                            Remove a worktree (keeps branch)
```

Detailed usage for each command follows below.

### wt init

**Usage:** `wt init`

Creates a `.wtconfig` template in the repository root with commented examples for common setups (Node.js, Python, Rust). The file is used by `wt new` and `wt eject` to automatically copy or symlink files into new worktrees.

**Note:** Only creates the file if it doesn't already exist.

**Default template:**

```
# .wtconfig — files to copy or symlink into new worktrees
# Syntax: <action> <path>
# Actions: copy, symlink

# Node.js
# copy .env.local
# symlink node_modules

# Python
# copy .env
# symlink .venv

# Rust
# symlink target
```

**How it works:** Uncomment or add lines to define which files should be automatically handled when creating new worktrees. For example, `copy .env.local` will copy your environment file into each new worktree, while `symlink node_modules` will create a symlink to share dependencies.

### wt new

**Usage:** `wt new [--copy <path>] [--symlink <path>] <name> [branch]`

Creates a new git worktree as a sibling directory. The worktree directory is named `<repo>-<name>` and placed alongside the main repo directory.

**Arguments:**
- `<name>` (required): Short name for the worktree. Used as directory suffix and default branch name.
- `[branch]` (optional): Branch name to use. Defaults to `<name>`. If the branch already exists, it checks it out; otherwise creates a new branch.

**Flags:**
- `--copy <path>`: Copy the specified file or directory from the source worktree instead of following .wtconfig default. Can be repeated.
- `--symlink <path>`: Symlink the specified file or directory from the source worktree instead of following .wtconfig default. Can be repeated.

**Behavior details:**
- Runs `_wt_setup` after creating the worktree, which processes `.wtconfig` (if it exists) and applies any `--copy`/`--symlink` overrides.
- Flag overrides take precedence over `.wtconfig` entries for the same path.
- Flags can also specify paths NOT in `.wtconfig` for one-off operations.
- After creation, automatically `cd`s into the new worktree.

**Examples:**

```zsh
# Create worktree with new branch "feature-auth"
wt new feature-auth

# Create worktree with existing branch
wt new hotfix release/hotfix-1.2

# Override .wtconfig: copy .env instead of symlinking
wt new feature-auth --copy .env

# One-off: symlink node_modules (even if not in .wtconfig)
wt new feature-auth --symlink node_modules

# Multiple overrides
wt new feature-auth --copy .env --symlink node_modules
```

### wt eject

**Usage:** `wt eject [--copy <path>] [--symlink <path>] [name]`

Moves the current branch out of the current worktree into a new dedicated worktree. Useful when you started work on a branch in the main worktree and want to isolate it.

**Arguments:**
- `[name]` (optional): Name for the new worktree directory. Defaults to the branch name with `/` replaced by `-`.

**Flags:**
- `--copy <path>`: Copy the specified file or directory from the source worktree instead of following .wtconfig default. Can be repeated.
- `--symlink <path>`: Symlink the specified file or directory from the source worktree instead of following .wtconfig default. Can be repeated.

**Behavior details:**
- Stashes any uncommitted changes (including untracked files).
- Switches the current worktree back to a fallback branch (main/master for home worktree, or the worktree's original branch for non-home worktrees).
- Creates a new worktree for the ejected branch.
- Restores stashed changes in the new worktree.
- Processes `.wtconfig` and applies flag overrides (same as `wt new`).
- After ejection, automatically `cd`s into the new worktree.

**Examples:**

```zsh
# Eject current branch into its own worktree
wt eject

# Eject with a custom worktree name
wt eject my-feature

# Eject with file overrides
wt eject --copy .env.local --symlink node_modules
```

## Tab Completion

wt includes built-in zsh tab completion. After sourcing `wt.zsh`, press Tab after `wt` to see available commands. Press Tab after `wt goto`, `wt merge`, `wt rebase`, or `wt delete` to complete worktree names.

The `wt new` and `wt eject` commands also complete `--copy` and `--symlink` flags with file paths from your current directory.

**Setup:** Completion is automatically registered via `compdef` when `wt.zsh` is sourced. No additional configuration needed.
