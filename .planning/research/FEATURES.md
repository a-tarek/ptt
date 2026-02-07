# Feature Landscape

**Domain:** Git worktree manager CLI tool (zsh to Go rewrite with npm distribution)
**Researched:** 2026-02-07

## Table Stakes

Features users expect from a modern Go CLI tool distributed via npm. Missing = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Shell completion generation | Modern CLI tools auto-generate completions for bash/zsh/fish | Low | Cobra provides built-in `completion` subcommand |
| `--help` flag on all commands | Universal CLI convention | Low | Cobra auto-generates from command descriptions |
| `--version` flag | Users need to verify installed version | Low | Standard root-level flag |
| Platform-specific binaries | npm package must work on macOS/Linux/WSL | Medium | Requires architecture-specific packages or postinstall detection |
| Exit codes (0=success, 1=error) | Scripts and CI depend on proper exit codes | Low | Already implemented in zsh version |
| Colored output with NO_COLOR support | Expected for modern CLIs, but must respect NO_COLOR env var | Low | Use library like fatih/color or charm/lipgloss |
| Error messages to stderr | stdout for data, stderr for errors | Low | Standard Go practice |
| Interactive installer | npm postinstall script that detects shell and offers to modify rc files | Medium | Must detect bash/zsh/fish, locate rc files, add source line |
| Shell wrapper generation | Commands that cd (goto, home, new, eject) need shell wrappers | High | Go binary generates shell function, sourced by rc file |
| Config file parsing (.wtconfig) | Existing feature, users depend on it | Medium | Simple line-based format: `<action> <path>` |
| Worktree name resolution | Existing feature: suffix matching on directory basename | Low | Port existing `_wt_resolve_path` logic |
| Subcommand architecture | 9 commands organized as subcommands | Low | Cobra command tree |
| Non-interactive mode | Support CI/scripting with no TTY | Low | Detect TTY, skip colored output/prompts |

## Differentiators

Features that set wt apart from alternatives. Not expected, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Automatic shell detection | Zero-config experience: installer detects shell and configures appropriately | Medium | Detect from SHELL env var, verify rc file exists |
| Override flags (--copy/--symlink) | Per-command overrides for .wtconfig behavior | Low | Already implemented, port to Go flags |
| Smart eject with stash handling | Safely move current branch to new worktree with uncommitted changes | High | Complex flow: stash, checkout, worktree add, stash pop |
| Suffix-based worktree matching | Type "staging" instead of "myapp-staging" | Low | Quality-of-life feature that differentiates from raw git worktree |
| Branch fallback detection | Eject from home → main/master, eject from worktree → inferred branch | Medium | Logic already exists in zsh, port to Go |
| Single binary + shell wrapper | Install via npm, works across all shells | High | Unique approach: binary does work, thin wrapper handles cd |
| .wtconfig templates | `wt init` generates commented examples for Node/Python/Rust | Low | Template embedding in Go binary |
| Worktree list with current marker | Show "*" next to current worktree in list output | Low | Parse git worktree list --porcelain |
| Config-free operation | Works without .wtconfig, just creates worktrees | Low | .wtconfig is optional enhancement |

## Anti-Features

Features to explicitly NOT build in the Go rewrite. Common mistakes in this domain.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Built-in cd implementation | Cannot cd from child process in Unix | Generate shell wrapper that sources a function |
| Complex config file format (YAML/TOML) | .wtconfig is intentionally simple for easy hand-editing | Keep line-based `<action> <path>` format |
| Automatic rc file modification | Don't modify user's shell config without permission | Installer offers, shows command, requires confirmation |
| GUI installer | npm postinstall runs in terminal, GUI adds complexity | Terminal-based interactive installer with clear prompts |
| Git subcommand (`git wt`) | Requires git config modification, harder to discover | Standalone `wt` command, simpler mental model |
| Worktree templates beyond copy/symlink | Feature creep: hooks, custom scripts, etc. | Stick to copy/symlink, users can script if needed |
| Interactive TUI for worktree management | Adds dependency, most operations are one-liners | Keep command-focused, list is already informative |
| Workspace persistence (remembering last worktree) | Adds state management, conflicts with shell history | Let shell history handle command recall |
| Auto-update mechanism | npm handles updates, don't duplicate | Use npm's standard update flow |
| Windows native support | WSL is sufficient for Git workflows | Document "use WSL on Windows" |

## Feature Dependencies

```
Platform-specific binaries
  └─→ Interactive installer (detects platform for wrapper syntax)
       └─→ Shell wrapper generation (bash/zsh/fish syntax differs)
            └─→ cd commands (goto, home, new, eject)

Config file parsing (.wtconfig)
  └─→ Override flags (--copy/--symlink)
       └─→ Setup function (_wt_setup equivalent)

Worktree name resolution
  └─→ Commands that accept worktree names (goto, merge, rebase, delete)

Shell completion generation
  └─→ Completion for worktree names (dynamic, queries git worktree list)
```

## MVP Recommendation

For Go rewrite MVP, prioritize:

1. **All 9 commands ported** - Core functionality, table stakes
2. **Shell wrapper generation** - Required for cd commands, core value prop
3. **Interactive installer** - npm postinstall, detects shell, offers to configure
4. **Shell completion (cobra built-in)** - Table stakes for modern CLI
5. **.wtconfig parsing** - Existing users depend on this
6. **Override flags** - Existing users depend on this
7. **Worktree name resolution** - Quality-of-life feature that's core to UX

Defer to post-MVP:
- **Colored output**: Low priority, focus on correctness first
- **Advanced error messages**: Can iterate after core works
- **Non-interactive mode flags**: Add when CI users request it

## Command Categorization

### Commands that require shell wrapper (cd changes parent shell)
- `goto <worktree>` - cd into worktree
- `home` - cd into main worktree
- `new [flags] <name> [branch]` - creates worktree, then cd into it
- `eject [flags] [name]` - creates worktree, then cd into it

**Implementation:** Go binary prints `CD:<path>` to stdout, shell wrapper parses and cd's.

### Commands that work as pure Go binary
- `init` - creates .wtconfig file
- `list` - queries git, prints to stdout
- `merge <worktree>` - calls `git merge`
- `rebase <worktree>` - calls `git rebase`
- `delete <worktree>` - calls `git worktree remove`

**Implementation:** Standard Cobra commands, no wrapper involvement.

## Shell Completion Requirements

| Completion Type | Commands | Source |
|----------------|----------|--------|
| Subcommand names | (root) | Static: new, goto, home, init, eject, list, merge, rebase, delete |
| Worktree names | goto, merge, rebase, delete | Dynamic: query `git worktree list --porcelain`, extract short names |
| Flag names | new, eject | Static: --copy, --symlink |
| File paths | --copy, --symlink arguments | File system completion (shell default) |
| Branch names | new [branch] argument | Dynamic: query `git branch`, but optional |

**Cobra implementation:** Custom completion function for worktree names, file completion for paths.

## npm Distribution Architecture

### Package Structure

```
@user/wt/
├── package.json (platform: neutral, postinstall script)
├── bin/
│   └── wt (Node.js shim that calls platform binary)
├── install.js (postinstall: detect platform, install wrapper)
├── binaries/
│   ├── wt-darwin-amd64
│   ├── wt-darwin-arm64
│   ├── wt-linux-amd64
│   └── wt-linux-arm64
└── wrappers/
    ├── wt.bash
    ├── wt.zsh
    └── wt.fish
```

**Alternative: Platform-specific packages**
```
@user/wt (depends on @user/wt-{platform})
@user/wt-darwin-arm64 (optionalDependencies in main package)
@user/wt-linux-amd64
...
```

**Recommendation:** Single package with all binaries, simpler for users.

## Shell Wrapper Pattern

### Wrapper Responsibilities
1. Source the shell function `wt()` into user's shell
2. Call Go binary with all arguments
3. Parse stdout for `CD:<path>` directive
4. Execute cd if directive present
5. Pass through all other output

### Example (bash/zsh):
```bash
wt() {
  local output
  output=$(/path/to/wt-binary "$@")
  local exit_code=$?

  if [[ "$output" == CD:* ]]; then
    local target="${output#CD:}"
    cd "$target"
  else
    echo "$output"
  fi

  return $exit_code
}
```

### Example (fish):
```fish
function wt
  set output (wt-binary $argv)
  set exit_code $status

  if string match -q 'CD:*' -- $output
    set target (string replace 'CD:' '' -- $output)
    cd $target
  else
    echo $output
  end

  return $exit_code
end
```

## Interactive Installer Flow

1. **Detect shell**: Read `$SHELL` env var → `/bin/zsh`, `/bin/bash`, `/usr/bin/fish`
2. **Locate rc file**:
   - bash: `~/.bashrc` (Linux) or `~/.bash_profile` (macOS)
   - zsh: `~/.zshrc`
   - fish: `~/.config/fish/config.fish`
3. **Check if already installed**: Grep rc file for `source.*wt.{shell}`
4. **Prompt user**: "Add wt to ~/.zshrc? [Y/n]"
5. **Show what will be added**:
   ```bash
   # wt - Git Worktree Manager
   source /path/to/node_modules/@user/wt/wrappers/wt.zsh
   ```
6. **Append if confirmed**
7. **Remind user**: "Run `source ~/.zshrc` or restart shell"

**Safety:**
- Never modify without confirmation
- Show exact line being added
- Detect existing installation (idempotent)
- Provide manual instructions if automated fails

## Configuration File Format (.wtconfig)

**Current format (keep this):**
```
# Comments start with #
<action> <path>

# Actions: copy, symlink
copy .env.local
symlink node_modules
```

**Parsing rules:**
- Line-based (split on newlines)
- Ignore lines starting with `#` (after trimming whitespace)
- Ignore blank lines
- Split on first space: `action` and `path`
- Validate action is "copy" or "symlink"
- Path is relative to repo root

**Go implementation:**
```go
type WtConfigEntry struct {
    Action string // "copy" or "symlink"
    Path   string // relative path
}

func ParseWtConfig(repoRoot string) ([]WtConfigEntry, error)
```

## Override Flags Design

**Current zsh implementation:**
```bash
wt new --copy .env.local --symlink node_modules myfeature
wt eject --copy .env --symlink target staging
```

**Behavior:**
- Override affects only the specified path for this command
- If path is in .wtconfig, override replaces the action
- If path is NOT in .wtconfig, override adds it as one-off
- Multiple `--copy` and `--symlink` flags allowed

**Go implementation:**
```go
// Cobra flags (repeatable)
cmd.Flags().StringArray("copy", []string{}, "Copy path (override .wtconfig)")
cmd.Flags().StringArray("symlink", []string{}, "Symlink path (override .wtconfig)")

// Merge logic
func MergeOverrides(configEntries []WtConfigEntry, copyPaths, symlinkPaths []string) []WtConfigEntry
```

## Complexity Assessment

| Feature Category | Overall Complexity | Risk Areas |
|-----------------|-------------------|------------|
| Core commands (port from zsh) | Low-Medium | eject is complex, rest are straightforward |
| Shell wrapper generation | Medium | Must handle 3 shell syntaxes correctly |
| Interactive installer | Medium | Shell detection, rc file modification edge cases |
| npm distribution | Low-Medium | Platform detection, binary selection |
| Config parsing | Low | Simple format, no dependencies |
| Completions | Low | Cobra generates boilerplate, custom for worktree names |
| Worktree resolution | Low | Port existing suffix-match logic |

## Edge Cases to Handle

| Edge Case | How to Handle |
|-----------|---------------|
| No .wtconfig file | Skip setup, create worktree only |
| .wtconfig references missing file | Log warning, skip that entry |
| User declines installer | Exit gracefully, show manual instructions |
| RC file doesn't exist | Create it (with user permission) or show manual instructions |
| Platform not detected | Error with manual installation instructions |
| Multiple shell rc files | Ask user which to configure |
| Already installed (rc file has source line) | Skip, report "already installed" |
| npm installed locally vs globally | Installer uses `__dirname` to find wrappers (works for both) |
| Worktree name collision | Same as zsh: error if target directory exists |
| Detached HEAD in eject | Same as zsh: error "nothing to eject" |

## Feature Parity Checklist

Port from zsh:
- [ ] 9 commands: new, goto, home, init, eject, list, merge, rebase, delete
- [ ] .wtconfig parsing (copy, symlink actions)
- [ ] Override flags (--copy, --symlink)
- [ ] Worktree name resolution (suffix matching)
- [ ] Eject with stash handling
- [ ] Branch fallback detection (main/master or inferred)
- [ ] List with current worktree marker (*)
- [ ] Error messages for all failure modes

Add for Go/npm:
- [ ] Shell wrapper generation (bash, zsh, fish)
- [ ] Interactive installer (npm postinstall)
- [ ] Platform binary selection
- [ ] Shell completion generation (cobra)
- [ ] --help and --version flags
- [ ] Proper exit codes
- [ ] CD directive protocol (binary → wrapper communication)

## Sources

**CONFIDENCE: MEDIUM to LOW**

This research is based on:
- **HIGH confidence:** Analysis of existing wt.zsh implementation (read directly from codebase)
- **MEDIUM confidence:** General knowledge of Cobra CLI framework patterns (from training data, pre-2025)
- **MEDIUM confidence:** npm binary distribution patterns (from training data, pre-2025)
- **LOW confidence:** Current best practices for Go CLI tools in 2026 (could not verify with WebSearch)
- **LOW confidence:** Specific Cobra completion API (could not verify with Context7 or official docs)

**Verification needed:**
- Cobra shell completion API (current version, 2026)
- npm best practices for platform-specific binaries (verify with official npm docs)
- Fish shell function syntax (verify wrapper example)
- Current conventions for NO_COLOR support (verify with no-color.org or similar)

**Note to roadmap creator:** Treat shell wrapper pattern and installer flow as HIGH confidence (well-established patterns). Treat specific library APIs (Cobra completion, npm distribution) as MEDIUM-LOW confidence pending verification with official documentation during implementation.
