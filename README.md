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
ptt init                  # Set up repo with .pttconfig/default.yml
ptt mk feature-auth       # Create worktree + branch, cd into it
# ... work on your feature ...
ptt cd                    # Jump back to main worktree
ptt cd feature-auth       # Jump to feature worktree
ptt rm feature-auth       # Remove worktree (keeps branch)
```

## Commands

| Command | Description |
|---------|-------------|
| `ptt init` | Create `.pttconfig/default.yml` template |
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

`ptt init` creates `.pttconfig/default.yml` at your repo root. It defines what happens when a new worktree is created or removed:

```yaml
# create:
#   - copy: .env
#   - symlink: node_modules
#   - run: npm install
#   - copyEnv:
#       file: .env
#       vars:
#         PORT:
#           strategy: ptt_increment
#         BRANCH:
#           strategy: git branch --show-current
#
# remove:
#   - run: echo "cleaning up"
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

## Prompt Integration

`ptt prompt` outputs a worktree-aware prompt segment: `🥔 [main +2 ~1]` inside a worktree, `🥔 root` at the bare repo root. Outside a ptt repo it exits with code 1, so you can fall back to your default git prompt.

Setup is opt-in — copy the snippet for your shell into your config.

### Zsh (Oh My Zsh / theme file)

Add this function to your theme file (e.g. `~/.oh-my-zsh/custom/themes/yourtheme.zsh-theme`),
then replace `$(git_prompt_info)` with `$(ptt_or_git)` in your `PROMPT`:

```zsh
ptt_or_git() {
  local out
  out=$(ptt prompt 2>/dev/null) && echo " $out" && return
  git_prompt_info
}
```

Before: `PROMPT='%~$(git_prompt_info) ❯ '`
After:  `PROMPT='%~$(ptt_or_git) ❯ '`

Your `ZSH_THEME_GIT_PROMPT_*` variables still work — they apply to the fallback path.

### Zsh (Oh My Zsh / zero-config override)

For automatic integration with any OMZ theme, add this to your `.zshrc` **after**
the `source $ZSH/oh-my-zsh.sh` line. No theme file changes needed:

```zsh
# Override git_prompt_info with ptt-aware version
if (( $+functions[git_prompt_info] )); then
  functions[_ptt_orig_git_prompt_info]=$functions[git_prompt_info]
  git_prompt_info() {
    local out
    out=$(ptt prompt 2>/dev/null) && echo " $out" && return
    _ptt_orig_git_prompt_info
  }
fi
```

This saves the original `git_prompt_info`, replaces it with one that tries
`ptt prompt` first, and falls back to the original for non-ptt repos.

### Zsh (no Oh My Zsh)

```zsh
ptt_or_git() {
  local out
  out=$(ptt prompt 2>/dev/null) && echo " $out" && return
  local branch=$(git branch --show-current 2>/dev/null)
  [ -n "$branch" ] && echo " ($branch)"
}
PROMPT='%~$(ptt_or_git)
❯ '
```

### Bash

```bash
ptt_or_git() {
  local out
  out=$(ptt prompt 2>/dev/null) && echo " $out" && return
  local branch=$(git branch --show-current 2>/dev/null)
  [ -n "$branch" ] && echo " ($branch)"
}
PS1='\w$(ptt_or_git)\n\$ '
```

### Fish

```fish
function ptt_or_git
  set -l out (ptt prompt 2>/dev/null)
  if test $status -eq 0; echo " $out"; return; end
  set -l branch (git branch --show-current 2>/dev/null)
  test -n "$branch" && echo " ($branch)"
end
```

## Uninstall

```bash
ptt uninstall
npm uninstall -g @atarek/ptt
```

## License

MIT
