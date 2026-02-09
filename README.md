<p align="center">
  <img src="assets/banner.svg" alt="ptt — git worktree manager" width="600">
</p>

<p align="center">
  A fast, cross-platform git worktree manager for bash, zsh, and fish.
</p>

<p align="center">
  <a href="https://www.npmjs.com/package/@atarek/ptt"><img src="https://img.shields.io/npm/v/@atarek/ptt" alt="npm"></a>
  <img src="https://img.shields.io/badge/platform-linux%20%7C%20macOS-blue" alt="platform">
  <img src="https://img.shields.io/badge/shell-bash%20%7C%20zsh%20%7C%20fish-green" alt="shells">
</p>

---

## Install

```bash
npm install -g @atarek/ptt
ptt install
```

Or try without installing: `npx @atarek/ptt install`

Requires Node.js 18+ and git 2.5+.

## Quick Start

```bash
ptt init                  # Create .pttconfig/default in your repo
ptt mk feature-auth       # Create worktree + branch, cd into it
# ... work on your feature ...
ptt cd                    # Jump back to main worktree
ptt cd feature-auth       # Jump to feature worktree
ptt rm feature-auth       # Remove worktree (keeps branch)
```

## Commands

| Command | Description |
|---------|-------------|
| `ptt init` | Create `.pttconfig/default` template |
| `ptt mk <name> [branch]` | Create worktree and switch to it |
| `ptt cd [name]` | Navigate to worktree (no args = main) |
| `ptt ls` | List all worktrees |
| `ptt rm <name>` | Remove worktree (`--branch` to also delete branch) |
| `ptt eject [name]` | Eject current branch into its own worktree |
| `ptt merge <name>` | Merge worktree's branch into current branch |
| `ptt rebase <name>` | Rebase current branch onto worktree's branch |
| `ptt mk-bare-repo` | Convert clone to bare repo with nested worktrees |

All worktree name arguments support **suffix matching** — `ptt cd auth` matches `myapp-feature-auth`.

## Configuration

`ptt init` creates `.pttconfig/default` at your repo root. It defines what happens when a new worktree is created:

```
copy .env                 # Independent copy per worktree
copy .env.local
symlink node_modules      # Shared via symlink (saves disk space)
symlink .venv
run npm install           # Run command in new worktree
```

**Override on the fly:**

```bash
ptt mk feature --copy .env --symlink node_modules
ptt mk staging --config staging    # Use .pttconfig/staging
ptt mk quick --skip-config         # Skip all config actions
```

## Bare Repo Support

For a cleaner layout where all worktrees live inside one directory:

```bash
ptt mk-bare-repo
cd ../myapp-bare/main
ptt mk feature-auth       # Creates myapp-bare/feature-auth/
```

```
myapp-bare/
├── .bare/          # Git database
├── .git            # Pointer to .bare
├── .pttconfig/     # Shared config
├── main/           # Default worktree
└── feature-auth/   # Created by ptt mk
```

## Uninstall

```bash
ptt uninstall
npm uninstall -g @atarek/ptt
```

## License

MIT
