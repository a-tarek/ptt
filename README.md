# ptt

A fast, cross-platform git worktree manager.

Managing git worktrees should be effortless. `ptt` makes it easy to create, navigate, and configure worktrees with intuitive commands that work the same way in bash, zsh, and fish shells across Linux, macOS, and Windows WSL.

## Features

- Create and navigate worktrees with short commands
- Auto-configure new worktrees with .pttconfig (copy/symlink files)
- Eject branches into dedicated worktrees with stash handling
- Tab completion for worktree names in all shells
- Works with both regular and bare git repos
- Single binary with fast startup (~5ms)

## Installation

### Primary Method (Global Install)

```bash
npm install -g @a-tarek/ptt
ptt install
```

### Try Before Committing

```bash
npx @a-tarek/ptt install
```

### What `ptt install` Does

The `ptt install` command:
- Detects your shell (bash, zsh, or fish)
- Adds a thin shell wrapper to your RC file (~/.bashrc, ~/.zshrc, or ~/.config/fish/config.fish)
- Sets up tab completion for worktree names
- Migrates from v1 if found (comments out old source lines)

The shell wrapper is needed because a subprocess cannot change the parent shell's directory. This is the same pattern used by tools like zoxide and nvm.

### Requirements

- Node.js 18+
- git 2.5+ (for worktree support)

### Uninstall

```bash
ptt uninstall
npm uninstall -g @a-tarek/ptt
```

## Quick Start

Here's a typical workflow with wt:

```bash
# 1. Initialize configuration in your repo
wt init
# Creates .wtconfig template at repo root

# 2. Create a new worktree and branch
wt new feature-auth
# Creates worktree and automatically switches to it

# 3. Work on your feature...
# (make changes, commit, etc.)

# 4. Jump back to main worktree
wt home
# Returns to the main worktree

# 5. Clean up when done
wt delete feature-auth
# Removes the worktree (keeps the branch)
```

**What's happening:**
- `wt init` creates a configuration file where you define which files to copy or symlink
- `wt new` creates a new worktree alongside your main repo (not nested inside it)
- `wt goto` and `wt home` let you navigate between worktrees
- `wt delete` cleans up worktrees you no longer need

## Commands

### wt init

**Usage:** `wt init`

Creates a `.wtconfig` template in the repository root with commented examples for common setups.

```bash
wt init
```

This creates a file like:

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

Uncomment or add lines to define which files should be automatically handled when creating new worktrees.

**When to run:** Once per repository, before creating your first worktree.

**Note:** If `.wtconfig` already exists, this command does nothing.

### wt new

**Usage:** `wt new [flags] <name> [branch]`

Creates a new git worktree and automatically switches to it.

**Arguments:**
| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | Short name for the worktree (used as directory suffix and default branch name) |
| `[branch]` | No | Branch name to use (defaults to `<name>`) |

**Flags:**
| Flag | Description |
|------|-------------|
| `--config <name>` | Use `.wtconfig-<name>` instead of `.wtconfig` |
| `--skip-config` | Skip all config file actions |
| `--copy <path>` | Copy the specified file/directory (repeatable, overrides config) |
| `--symlink <path>` | Symlink the specified file/directory (repeatable, overrides config) |

**Examples:**

```bash
# Create worktree with new branch "feature-auth"
wt new feature-auth

# Create worktree using existing branch
wt new hotfix release/hotfix-1.2

# Override .wtconfig: copy .env instead of symlinking
wt new feature-auth --copy .env

# One-off symlink (even if not in .wtconfig)
wt new feature-auth --symlink node_modules

# Use alternate config file
wt new staging --config staging
# Uses .wtconfig-staging instead of .wtconfig
```

**Behavior:**
- Creates worktree as sibling directory (e.g., `myapp-feature-auth` next to `myapp`)
- For bare repos, creates worktree inside the repo directory
- Processes `.wtconfig` (if it exists) to copy/symlink files
- Flag overrides merge with file-based config
- Automatically switches to the new worktree after creation

### wt goto

**Usage:** `wt goto <worktree>`

Navigate to a worktree by name.

**Arguments:**
| Argument | Required | Description |
|----------|----------|-------------|
| `<worktree>` | Yes | Worktree name (suffix matching) |

**Examples:**

```bash
wt goto feature-auth
# Matches "myapp-feature-auth"

wt goto staging
# Matches "myapp-staging"
```

**Behavior:**
- Name matching uses suffix matching (e.g., "staging" matches "myapp-staging")
- If already in the target worktree, prints a message and exits successfully
- Tab completion shows available worktree names

### wt home

**Usage:** `wt home`

Navigate back to the main (first) worktree.

```bash
wt home
```

**Behavior:**
- Finds the home worktree (the first one listed by `git worktree list`)
- If already in the home worktree, prints a message and exits successfully

### wt eject

**Usage:** `wt eject [flags] [name]`

Eject the current branch into its own dedicated worktree. Useful when you started work in the main worktree and want to isolate it.

**Arguments:**
| Argument | Required | Description |
|----------|----------|-------------|
| `[name]` | No | Name for the new worktree (defaults to branch name with `/` replaced by `-`) |

**Flags:**
| Flag | Description |
|------|-------------|
| `--config <name>` | Use `.wtconfig-<name>` instead of `.wtconfig` |
| `--skip-config` | Skip all config file actions |
| `--copy <path>` | Copy the specified file/directory (repeatable, overrides config) |
| `--symlink <path>` | Symlink the specified file/directory (repeatable, overrides config) |

**Examples:**

```bash
# Eject current branch into its own worktree
wt eject

# Eject with a custom worktree name
wt eject my-feature

# Eject with file overrides
wt eject --copy .env.local --symlink node_modules
```

**Behavior:**
- Stashes any uncommitted changes (including untracked files)
- Switches current worktree back to a fallback branch:
  - For home worktree: switches to main or master
  - For other worktrees: switches to the original branch for that worktree
- Creates new worktree for the ejected branch
- Restores stashed changes in the new worktree
- Processes `.wtconfig` and applies flag overrides
- Automatically switches to the new worktree

**Note:** If stash pop has merge conflicts, a warning is printed but the command succeeds. You can resolve conflicts in the new worktree.

### wt list

**Usage:** `wt list`

List all worktrees with their directory names and branches.

```bash
wt list
```

**Example output:**

```
* myapp                          main
  myapp-feature-auth             feature-auth
  myapp-staging                  staging
```

The `*` marks the current worktree.

### wt merge

**Usage:** `wt merge <worktree>`

Merge the specified worktree's branch into the current branch.

**Arguments:**
| Argument | Required | Description |
|----------|----------|-------------|
| `<worktree>` | Yes | Worktree name (suffix matching) |

**Examples:**

```bash
# From main worktree, merge feature-auth's branch
wt merge feature-auth

# Equivalent to:
# git merge feature-auth
```

**Behavior:**
- Resolves the worktree name to its branch
- Runs `git merge <branch>`
- Tab completion shows available worktree names

### wt rebase

**Usage:** `wt rebase <worktree>`

Rebase the current branch onto the specified worktree's branch.

**Arguments:**
| Argument | Required | Description |
|----------|----------|-------------|
| `<worktree>` | Yes | Worktree name (suffix matching) |

**Examples:**

```bash
# From feature branch, rebase onto main
wt rebase main

# Equivalent to:
# git rebase main
```

**Behavior:**
- Resolves the worktree name to its branch
- Runs `git rebase <branch>`
- Tab completion shows available worktree names

### wt delete

**Usage:** `wt delete [flags] <worktree>`

Remove a worktree directory. The branch is kept — only the worktree checkout is removed.

**Arguments:**
| Argument | Required | Description |
|----------|----------|-------------|
| `<worktree>` | Yes | Worktree name (suffix matching) |

**Flags:**
| Flag | Description |
|------|-------------|
| `--branch` | Also delete the branch after removing the worktree |

**Examples:**

```bash
# Remove worktree, keep branch
wt delete feature-auth

# Remove worktree and branch
wt delete feature-auth --branch
```

**Behavior:**
- Uses `git worktree remove` to remove the worktree
- If worktree has uncommitted changes (dirty), prompts for confirmation
- Clean worktrees are removed silently
- With `--branch` flag, also deletes the branch after removing the worktree
- Tab completion shows available worktree names

### ptt install

**Usage:** `ptt install`

Interactive installer that sets up ptt shell integration.

```bash
ptt install
```

**What it does:**
- Detects your shell (bash, zsh, or fish)
- Checks for existing ptt installations
- Shows you exactly what will be added to your RC file
- Migrates from v1 if found (comments out old source lines)
- Creates a backup of your RC file before modifying
- Adds the ptt configuration block with markers

**The installation is idempotent** — you can run it multiple times safely. If ptt is already installed, it will detect this and exit without making changes.

### ptt uninstall

**Usage:** `ptt uninstall`

Removes ptt shell integration from your RC file.

```bash
ptt uninstall
```

**What it does:**
- Detects your shell
- Removes the ptt configuration block from your RC file
- Creates a backup before modifying
- Prints instructions for removing the npm package

**Note:** This command only removes the shell integration. To completely remove ptt, also run:
```bash
npm uninstall -g @a-tarek/ptt
```

## Configuration

### .wtconfig

The `.wtconfig` file lives at your repository root and defines which files to copy or symlink when creating new worktrees.

**Syntax:**

```
# Comment lines start with #
<action> <path>
```

**Actions:**
- `copy` — Duplicate the file or directory from the source worktree. Each worktree gets its own independent copy.
- `symlink` — Create a symbolic link to the source file or directory. Changes affect all worktrees.

**Full example for a Node.js project:**

```
# .wtconfig — files to copy or symlink into new worktrees

# Environment files - copy so each worktree can have different settings
copy .env
copy .env.local
copy .env.test

# Dependencies - symlink to save disk space
symlink node_modules

# Build cache - symlink for faster builds
symlink .next
```

### When to Use Copy vs Symlink

**Use `copy` for:**
- Environment variables (.env files) that differ per worktree
- Configuration files you want to customize per worktree
- Files that might be modified during development
- Docker compose files with different port mappings

**Use `symlink` for:**
- Large dependency directories (node_modules, .venv, target)
- Build caches that can be shared
- Static files that never change
- Read-only assets

**Example decision tree:**

| File/Directory | Action | Reason |
|----------------|--------|--------|
| `.env` | copy | Different environment per worktree |
| `node_modules` | symlink | Saves disk space, dependencies rarely differ |
| `.env.local` | copy | Per-worktree local overrides |
| `.venv` | symlink | Python virtual environment shared |
| `target` | symlink | Rust build cache shared |
| `.next` | symlink | Next.js build cache shared |
| `docker-compose.override.yml` | copy | Different ports per worktree |

### Alternate Config Files

You can create named config files like `.wtconfig-staging` or `.wtconfig-ci` for different scenarios:

```bash
# Use default .wtconfig
wt new feature-auth

# Use .wtconfig-staging
wt new staging --config staging

# Use .wtconfig-ci
wt new ci-test --config ci
```

**When to use:**
- Different dependency setups for different environments
- CI/staging worktrees need different file handling
- Production worktrees have stricter requirements

### Override Flags

The `--copy` and `--symlink` flags on `wt new` and `wt eject` let you override `.wtconfig` defaults on a per-command basis.

**Precedence:** CLI flags > .wtconfig entries

**Example scenarios:**

```bash
# .wtconfig says "symlink .env" but you need a separate copy
wt new feature-auth --copy .env

# .wtconfig doesn't mention node_modules, but symlink it this time
wt new feature-auth --symlink node_modules

# Multiple overrides
wt new feature-auth --copy .env --copy .env.local --symlink node_modules
```

**How it works:**
1. `.wtconfig` entries are loaded and processed
2. For paths that appear in both config and flags, flags take precedence
3. Paths that only appear in flags are applied as one-off operations

This lets you maintain a standard configuration while making exceptions when needed.

## Shell Support

### Supported Shells

| Shell | Min Version | Completion Support |
|-------|-------------|-------------------|
| bash  | 3.2+        | Yes               |
| zsh   | 5.8+        | Yes               |
| fish  | 3.0+        | Yes               |

### How It Works

The `ptt` command consists of two parts:

1. **Go binary** (`ptt`) — Handles all logic, git operations, and completion generation
2. **Shell wrapper** — Thin function that handles directory changes

When you run `ptt go feature-auth`, the wrapper:
- Calls the Go binary with `--output-path` flag
- Binary outputs the target path to stdout
- Wrapper uses that path to `cd` into the directory

For non-directory-changing commands (like `ptt ls`, `ptt merge`), the wrapper simply passes through to the binary.

This pattern is used by tools like zoxide, nvm, and direnv because a subprocess cannot change the parent shell's directory.

### Tab Completion

Tab completion is automatically enabled for commands that take worktree names as arguments:

```bash
ptt go <TAB>         # Shows worktree names
ptt rm <TAB>         # Shows worktree names
ptt merge <TAB>      # Shows worktree names
ptt rebase <TAB>     # Shows worktree names
```

The completion is **live** — it queries `git worktree list` on every tab press to ensure accuracy. This adds ~5-10ms latency, which is acceptable for the improved UX.

Commands that create new names (`ptt mk`, `ptt eject`) do not have completion for the name argument.

## Troubleshooting

### "ptt: command not found"

**Cause:** The shell wrapper hasn't been sourced yet.

**Solution:**
```bash
# Run the installer
ptt install

# Then either restart your terminal or source your RC file:
source ~/.bashrc    # for bash
source ~/.zshrc     # for zsh
source ~/.config/fish/config.fish  # for fish
```

### "not inside a git repository"

**Cause:** You're running ptt commands outside a git repository.

**Solution:** Navigate to a git repository before running ptt commands:
```bash
cd /path/to/your/git/repo
ptt ls
```

### Tab completion not working

**Cause:** Shell hasn't loaded the completion configuration yet.

**Solution:**
```bash
# Restart your terminal, or re-run the installer
ptt install
```

If tab completion still doesn't work after restarting:
- **bash**: Check that bash-completion is installed (`brew install bash-completion` on macOS)
- **zsh**: Check that compinit is called in your .zshrc
- **fish**: Completions should work automatically

### Worktree name not found

**Cause:** The name you provided doesn't match any worktree suffix.

**Solution:**
```bash
# Check available worktrees
ptt ls

# Use the suffix from the directory name
# If you see "myapp-feature-auth", use:
ptt go feature-auth
```

The matching is based on the directory name suffix. If your worktree is named `myapp-feature-auth`, you can navigate to it with `ptt go feature-auth` or just `ptt go auth`.

### Permission error during install

**Cause:** The npm global install requires elevated permissions.

**Solution:**
```bash
# Either use sudo (not recommended long-term)
sudo npm install -g @a-tarek/ptt

# Or configure npm to use a user directory (recommended)
mkdir ~/.npm-global
npm config set prefix '~/.npm-global'
echo 'export PATH=~/.npm-global/bin:$PATH' >> ~/.bashrc
source ~/.bashrc
npm install -g @a-tarek/ptt
```

See [npm documentation on fixing permissions](https://docs.npmjs.com/resolving-eacces-permissions-errors-when-installing-packages-globally).

### Different behavior between shells

**Cause:** Each shell has slight syntax differences in the wrapper function.

**Solution:** If you encounter shell-specific issues, file a bug report with:
- Your shell and version (`bash --version`, `zsh --version`, or `fish --version`)
- The exact command that failed
- The error message

The wrappers are designed to be functionally identical across shells.

## License

MIT
