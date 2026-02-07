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

## Tab Completion

wt includes built-in zsh tab completion. After sourcing `wt.zsh`, press Tab after `wt` to see available commands. Press Tab after `wt goto`, `wt merge`, `wt rebase`, or `wt delete` to complete worktree names.

The `wt new` and `wt eject` commands also complete `--copy` and `--symlink` flags with file paths from your current directory.

**Setup:** Completion is automatically registered via `compdef` when `wt.zsh` is sourced. No additional configuration needed.
