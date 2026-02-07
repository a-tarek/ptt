# Phase 6: Shell Integration - Research

**Researched:** 2026-02-07
**Domain:** Shell wrapper functions, Cobra completion system, multi-shell scripting (bash/zsh/fish)
**Confidence:** HIGH

## Summary

Shell integration for CLI tools that change directories requires a shell wrapper function to execute `cd` in the parent shell process (child processes cannot change their parent's working directory). The standard pattern, used by tools like zoxide, direnv, and starship, involves generating shell-specific initialization code via an `eval $(tool init shell)` command that users add to their rc files.

Cobra v1.10.2 provides built-in completion generation for bash, zsh, and fish via dedicated commands (`wt completion bash`). Dynamic completions are implemented using `ValidArgsFunction` which queries data (like git worktree list) on every tab press, ensuring accuracy without caching complexity.

The wrapper intercepts all `wt` commands, delegates to the binary with `--output-path` flag, captures stdout for cd coordination, and lets stderr pass through for user messages. The binary uses `go:embed` directives to bundle shell scripts, creating a single self-contained binary.

**Primary recommendation:** Use `wt shell-init` command with auto-detection to generate wrapper code; use Cobra's `ValidArgsFunction` for dynamic worktree name completions; embed wrapper scripts via `go:embed` for single-binary distribution; target bash 3.2+ for macOS compatibility.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Wrapper sourcing & setup:**
- `wt shell-init` command outputs wrapper code for the detected shell
- Auto-detects current shell (no explicit argument needed)
- User adds `eval $(wt shell-init)` to their rc file (or installer does it in Phase 8)
- Completions are a separate command: `wt completion <shell>` (two eval lines, not bundled)
- No customization flags (no `--cmd` alias, no `--no-alias`)
- Function is always named `wt`
- No legacy wt.zsh detection — shell-init is standalone, ignores existing wt.zsh
- No env vars set after cd

**cd command behavior:**
- On successful cd: show worktree info (name + branch + path) then change directory
- Errors follow existing Go binary error patterns (consistent with Phase 3-5)
- Current stdout/stderr protocol is kept: binary prints path to stdout via `--output-path`, info/errors to stderr, wrapper captures stdout for cd
- `wt new` shows setup summary after creation (config actions that ran + worktree info)
- `wt eject` auto-cd's into the new worktree after completion
- "Already there" case: just show message, no redundant cd
- `wt merge`/`wt rebase`: let git output pass through naturally, no wrapper formatting
- No environment variables set after cd

**Tab completion:**
- Use Cobra's built-in completion generation for all shells
- Dynamic completions: every tab press queries git worktree list (always accurate, no caching)
- Worktree name completions show names only (no branch descriptions)
- `wt new` has no completions for the branch name argument (it's a new name)

**Multi-shell parity:**
- All three shells (bash, zsh, fish) supported in this phase — no deferral
- Identical user-facing behavior across all shells — no shell-specific differences
- No shell mismatch validation in shell-init

### Claude's Discretion

- All wt commands route through wrapper vs only cd commands (leaning: all through wrapper — simpler mental model, like zoxide)
- Wrapper calls binary via `command wt` to bypass function (standard pattern)
- Wrapper scripts embedded in Go binary via `go:embed` (leaning: embed — single binary distribution, no missing files)
- Minimum bash version support (leaning: bash 3.2+ for macOS compatibility, wrapper is simple enough)
- Exact worktree info format on successful cd

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope

</user_constraints>

## Standard Stack

The established tools and patterns for shell integration:

### Core

| Tool/Pattern | Version | Purpose | Why Standard |
|--------------|---------|---------|--------------|
| Cobra | v1.10.2 | CLI framework with built-in completions | Industry standard (used by Kubernetes, Hugo, GitHub CLI); supports bash/zsh/fish/powershell completions natively |
| go:embed | Go 1.16+ | Embed files in binary | Official Go feature for bundling assets; eliminates separate file distribution |
| eval pattern | N/A | Shell initialization via `eval $(cmd init shell)` | Used by zoxide, direnv, starship, atuin; familiar to users; supports all shells |
| command builtin | POSIX | Bypass shell functions | Standard way to call binary from wrapper function with same name |

### Supporting

| Tool/Pattern | Version | Purpose | When to Use |
|--------------|---------|---------|-------------|
| $SHELL variable | POSIX | Detect current shell | Auto-detection in shell-init command |
| git worktree list --porcelain | Git 2.7+ | Query worktrees | Dynamic completion data source |
| ValidArgsFunction | Cobra 1.13+ | Dynamic completions | When completion data changes (worktree names) |
| ShellCompDirective | Cobra | Control completion behavior | Disable file completion, maintain order |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| go:embed | Ship .sh/.fish files separately | Easier to debug but adds install complexity; risk of missing files |
| Dynamic completion | Static completion + cache | Faster but stale data; added cache invalidation complexity |
| eval pattern | Manual function copy-paste | No auto-update but harder for users; harder to maintain |
| ValidArgsFunction | External completion file | Shell-specific syntax but more flexible for complex completions |

**Installation:**
```bash
# Already in project (Cobra)
# go:embed is built into Go 1.16+
```

## Architecture Patterns

### Recommended Project Structure

```
cmd/
├── shell_init.go         # wt shell-init command
├── completion.go         # wt completion <shell> command (Cobra built-in)
├── goto.go               # Already exists with --output-path
└── root.go               # Root command setup

internal/
└── shell/
    ├── templates/        # Embedded shell scripts
    │   ├── wrapper.bash
    │   ├── wrapper.zsh
    │   └── wrapper.fish
    └── detect.go         # Shell detection logic

# Cobra generates completion commands automatically
# No need for custom completion files
```

### Pattern 1: Shell Wrapper Coordination

**What:** Binary outputs target path on stdout when `--output-path` flag is present; wrapper captures it and executes `cd`. This pattern works because the wrapper function executes in the parent shell process, not a child process.

**When to use:** Any CLI tool that needs to change the shell's working directory (zoxide, autojump, z, etc.)

**Example:**
```bash
# Bash/Zsh wrapper pattern
wt() {
  local output
  output=$(command wt --output-path "$@")
  local exit_code=$?

  if [ $exit_code -eq 0 ] && [ -n "$output" ]; then
    cd "$output"
  fi

  return $exit_code
}
```

```fish
# Fish wrapper pattern
function wt
  set -l output (command wt --output-path $argv)
  set -l exit_code $status

  if test $exit_code -eq 0 -a -n "$output"
    cd "$output"
  end

  return $exit_code
end
```

**Key insight:** The `command` builtin/function ensures the wrapper calls the actual binary, not itself recursively. Source: [Bash bypass alias](https://www.cyberciti.biz/faq/bash-bypass-alias-command-on-linux-macos-unix/)

### Pattern 2: Shell Auto-Detection

**What:** Detect current shell from `$SHELL` environment variable to generate appropriate wrapper code.

**When to use:** When generating shell-specific initialization code via `eval $(tool init)` pattern.

**Example:**
```go
// Source: inspired by zoxide's approach
func DetectShell() (string, error) {
    shellPath := os.Getenv("SHELL")
    if shellPath == "" {
        return "", fmt.Errorf("SHELL environment variable not set")
    }

    basename := filepath.Base(shellPath)

    switch basename {
    case "bash":
        return "bash", nil
    case "zsh":
        return "zsh", nil
    case "fish":
        return "fish", nil
    default:
        return "", fmt.Errorf("unsupported shell: %s", basename)
    }
}
```

**Caveat:** Fish uses `fish` as $SHELL, bash uses `/bin/bash`, zsh uses `/bin/zsh` or `/usr/local/bin/zsh`. Extract basename only. Source: [Fish FAQ](https://fishshell.com/docs/current/faq.html)

### Pattern 3: Embed Shell Scripts with go:embed

**What:** Embed shell wrapper templates in Go binary at compile time using `//go:embed` directive.

**When to use:** CLI tools that need to distribute shell integration code without external files.

**Example:**
```go
package shell

import (
    _ "embed"
    "text/template"
)

//go:embed templates/wrapper.bash
var bashWrapper string

//go:embed templates/wrapper.zsh
var zshWrapper string

//go:embed templates/wrapper.fish
var fishWrapper string

func GenerateWrapper(shell string) (string, error) {
    switch shell {
    case "bash":
        return bashWrapper, nil
    case "zsh":
        return zshWrapper, nil
    case "fish":
        return fishWrapper, nil
    default:
        return "", fmt.Errorf("unsupported shell: %s", shell)
    }
}
```

Source: [Go embed package](https://pkg.go.dev/embed), [Using Go Embed Package](https://andrew-mccall.com/blog/2025/01/using-go-embed-package-for-template-rendering/)

### Pattern 4: Dynamic Cobra Completions

**What:** Use `ValidArgsFunction` to provide completions that query live data on every tab press.

**When to use:** When completion candidates change frequently (worktrees, branches, running processes, etc.)

**Example:**
```go
// Source: Cobra documentation
var gotoCmd = &cobra.Command{
    Use:   "goto <worktree>",
    Short: "Switch to a worktree",
    ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        // Query git worktree list on every tab press
        worktrees, err := git.ListWorktrees()
        if err != nil {
            return nil, cobra.ShellCompDirectiveError
        }

        // Extract names for completion
        names := make([]string, 0, len(worktrees))
        for _, wt := range worktrees {
            name := extractWorktreeName(wt.Path)
            names = append(names, name)
        }

        return names, cobra.ShellCompDirectiveNoFileComp
    },
}
```

**Key insight:** Don't cache worktree list. Query on every completion for accuracy. Performance is acceptable (git worktree list is fast). Source: [Cobra completions guide](https://cobra.dev/docs/how-to-guides/shell-completion/)

### Pattern 5: Separate Completion Command

**What:** Cobra provides automatic completion command generation. Users run `wt completion bash > ~/.bash_completions/wt` once, then completions work forever.

**When to use:** Standard approach for CLI tools using Cobra.

**Example:**
```go
// Cobra automatically adds this command
// User runs: wt completion bash
// Output: Complete bash completion script
// User installs: wt completion bash > ~/.bash_completions/wt

// In root.go:
func init() {
    // Cobra adds completion command automatically
    // Just need to ensure ValidArgsFunction is set on commands
}
```

Source: [Cobra shell completion](https://cobra.dev/docs/how-to-guides/shell-completion/), [Shell completions with Cobra](https://blog.chmouel.com/posts/cobra-completions/)

### Anti-Patterns to Avoid

- **Caching completion data:** Stale worktree names if cache isn't invalidated. Query live data instead.
- **Using same function name as binary without `command` bypass:** Infinite recursion.
- **Mixing shell-init and completion into single output:** Users can't opt into just wrapper or just completions. Keep separate.
- **Setting environment variables after cd:** Pollutes shell environment; makes wrapper stateful. Keep wrapper simple.
- **Validating shell mismatch (e.g., zsh wrapper in bash):** Over-engineering; shell-init will be called from correct shell in Phase 8 installer.
- **Bundling fish/zsh completion with wrapper:** Completions require separate installation (different paths, different activation). Two eval lines.

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Shell completion system | Custom completion parser | Cobra's ValidArgsFunction | Cross-shell compatibility already solved; zsh descriptions work automatically; fish syntax differences handled |
| Embedding assets in binary | Custom build script | go:embed directive | Official Go feature since 1.16; compile-time safety; no runtime overhead |
| Shell script template engine | String concatenation | text/template (if needed) or static files | Maintainability; but for simple wrappers, embedded strings are fine |
| Shell detection | Parse $0 or ps output | $SHELL environment variable | Standard POSIX approach; reliable across systems |
| Command bypass (calling binary from function) | $PATH manipulation | `command` builtin | POSIX standard; works in bash/zsh; fish has `command` as well |

**Key insight:** Cobra has solved cross-shell completion complexity. Don't reimplement completion scripts manually. Use ValidArgsFunction and Cobra's built-in completion commands.

## Common Pitfalls

### Pitfall 1: Bash 3.2 Compatibility on macOS

**What goes wrong:** Using modern bash features (associative arrays, `[[` with `==`, etc.) breaks on macOS default bash 3.2.

**Why it happens:** macOS ships bash 3.2 (from 2007) due to GPLv3 licensing. Version 4.0+ uses GPLv3 which Apple doesn't accept. Many users never upgrade.

**How to avoid:**
- Use `[ "$var" = "value" ]` not `[[ "$var" == "value" ]]` (test with `=` not `==`)
- Avoid associative arrays (`declare -A`). Use indexed arrays only.
- Avoid `readarray`/`mapfile` (bash 4+). Use `while read` loops.
- Test wrappers on macOS bash 3.2 explicitly.

**Warning signs:** Syntax errors on macOS but works on Linux bash 4+. Users report "command not found" on macOS.

Source: [Upgrading Bash on macOS](https://itnext.io/upgrading-bash-on-macos-7138bd1066ba), [Bash 3.2 compatibility issues](https://medium.com/macoclock/shell-bash-zsh-macos-catalina-145f378a8381)

### Pitfall 2: Fish Syntax Completely Different from Bash

**What goes wrong:** Copying bash wrapper to fish with minor tweaks fails. Fish uses fundamentally different syntax.

**Why it happens:** Fish doesn't aim for POSIX compatibility. Control structures, variables, and functions all use different keywords.

**How to avoid:**
- Recognize fish is not bash-like: `function ... end` not `func() { ... }`
- Use `set` for all variable assignments: `set -l var value` not `var=value`
- Use `test` or `[` for conditionals (no `[[`): `test $status -eq 0` not `[[ $? -eq 0 ]]`
- Command substitution: `(command)` not `$(command)` (both work but `()` is fish idiom)
- Variable expansion: `$var` not `${var}` (fish is simpler)
- Arrays: `set -l arr a b c` then `$arr[1]` not `${arr[0]}` (1-indexed!)

**Warning signs:** Fish reports "Unknown command 'var'" or "Unexpected end of string" for bash syntax.

Source: [Fish for bash users](https://fishshell.com/docs/current/fish_for_bash_users.html), [Fish function documentation](https://fishshell.com/docs/current/cmds/function.html)

### Pitfall 3: Recursive Function Calls

**What goes wrong:** Wrapper function calls itself instead of binary, causing infinite recursion or "command not found" errors.

**Why it happens:** Shell looks up function name before $PATH. Function `wt()` shadows binary `wt`.

**How to avoid:**
- Use `command wt` in bash/zsh wrapper (bypasses functions, calls binary from $PATH)
- Use `command wt` in fish wrapper too (fish has `command` builtin)
- Never call `wt` directly in the wrapper; always use `command wt`

**Warning signs:** Stack overflow, shell hangs, or "wt: command not found" when binary is definitely in $PATH.

Source: [Shell wrapper bypass](https://www.cyberciti.biz/faq/bash-bypass-alias-command-on-linux-macos-unix/), [Wrapping shell commands](https://kevinjalbert.com/wrapping-shell-commands-and-keep-the-original-name/)

### Pitfall 4: Completion Script Installation vs Wrapper Sourcing

**What goes wrong:** Confusing the wrapper function (must be eval'd in rc file) with completion scripts (must be written to specific directories).

**Why it happens:** Two different shell integration mechanisms serve different purposes but both involve "sourcing" in some form.

**How to avoid:**
- **Wrapper:** Must be `eval $(wt shell-init)` in rc file every shell startup. Defines function in current shell session.
- **Completion:** One-time install to shell completion directory (e.g., `/etc/bash_completion.d/`, `~/.config/fish/completions/`). Shell loads automatically.
- Cobra completion commands output full script to stdout; redirect to file, don't eval.
- Document both steps clearly: one eval line for wrapper, one redirect for completion.

**Warning signs:** Tab completion doesn't work (completion script not installed). Directory changes don't work (wrapper not eval'd).

Source: [Cobra completion installation](https://cobra.dev/docs/how-to-guides/shell-completion/)

### Pitfall 5: Capturing Stderr in Wrapper

**What goes wrong:** Wrapper captures stderr along with stdout, swallowing error messages.

**Why it happens:** Using `$()` captures stdout only by default, which is correct. But some scripts mistakenly use `2>&1` inside `$()`.

**How to avoid:**
- Only capture stdout: `output=$(command wt --output-path "$@")`
- Let stderr flow to terminal: don't redirect 2>&1 in wrapper
- Binary already prints user messages to stderr, which will display naturally

**Warning signs:** Users don't see error messages. Silent failures.

### Pitfall 6: Empty Output Treated as Success

**What goes wrong:** Wrapper runs `cd ""` when binary fails but exits with 0 and empty stdout.

**Why it happens:** Not checking both exit code AND non-empty output before cd.

**How to avoid:**
```bash
# CORRECT: Check exit code AND output is non-empty
if [ $exit_code -eq 0 ] && [ -n "$output" ]; then
  cd "$output"
fi
```

```bash
# WRONG: Only check exit code
if [ $exit_code -eq 0 ]; then
  cd "$output"  # Might be empty!
fi
```

**Warning signs:** Users report "cd: empty directory name" or shell cd's to home directory unexpectedly.

### Pitfall 7: ValidArgsFunction with ShellCompDirectiveDefault

**What goes wrong:** File completion appears alongside worktree names, confusing users.

**Why it happens:** `ShellCompDirectiveDefault` allows file completion as fallback. For worktree names, this is wrong.

**How to avoid:**
```go
// CORRECT: Disable file completion
return names, cobra.ShellCompDirectiveNoFileComp

// WRONG: Allows file completion
return names, cobra.ShellCompDirectiveDefault
```

**Warning signs:** Tab completion shows files in current directory mixed with worktree names.

Source: [Cobra shell directives](https://pkg.go.dev/github.com/spf13/cobra)

## Code Examples

Verified patterns from official sources:

### Shell Detection (Go)

```go
// Source: Inspired by zoxide's approach
package shell

import (
    "fmt"
    "os"
    "path/filepath"
)

func DetectShell() (string, error) {
    shellPath := os.Getenv("SHELL")
    if shellPath == "" {
        return "", fmt.Errorf("SHELL environment variable not set")
    }

    basename := filepath.Base(shellPath)

    switch basename {
    case "bash":
        return "bash", nil
    case "zsh":
        return "zsh", nil
    case "fish":
        return "fish", nil
    default:
        return "", fmt.Errorf("unsupported shell: %s (only bash, zsh, fish supported)", basename)
    }
}
```

### Bash/Zsh Wrapper Function

```bash
# Source: Adapted from zoxide pattern
# Works on bash 3.2+ (macOS compatible)
wt() {
  local output
  output=$(command wt --output-path "$@")
  local exit_code=$?

  if [ $exit_code -eq 0 ] && [ -n "$output" ]; then
    cd "$output"
  fi

  return $exit_code
}
```

**Key features:**
- `local output` declares local variable (bash 3.2+)
- `command wt` bypasses function, calls binary
- Checks exit code AND non-empty output
- Returns binary's exit code to shell

### Fish Wrapper Function

```fish
# Source: Adapted from zoxide pattern
function wt
  set -l output (command wt --output-path $argv)
  set -l exit_code $status

  if test $exit_code -eq 0 -a -n "$output"
    cd "$output"
  end

  return $exit_code
end
```

**Key features:**
- `set -l` creates local variable (fish idiom)
- `$argv` is fish's `$@` equivalent (all arguments)
- `$status` is fish's `$?` equivalent (exit code)
- `test -a` is AND operator (fish doesn't have `&&` in conditionals)
- `function ... end` syntax (not `func() { ... }`)

### Embedding Wrapper Scripts (Go)

```go
// Source: Go embed documentation
package shell

import (
    _ "embed"
)

//go:embed templates/wrapper.bash
var bashWrapper string

//go:embed templates/wrapper.zsh
var zshWrapper string

//go:embed templates/wrapper.fish
var fishWrapper string

func GetWrapper(shell string) (string, error) {
    switch shell {
    case "bash":
        return bashWrapper, nil
    case "zsh":
        return zshWrapper, nil
    case "fish":
        return fishWrapper, nil
    default:
        return "", fmt.Errorf("unsupported shell: %s", shell)
    }
}
```

### Shell-Init Command (Go)

```go
// Source: Cobra command pattern
package cmd

import (
    "fmt"

    "github.com/ahmedelarabyy/wt/internal/shell"
    "github.com/spf13/cobra"
)

var shellInitCmd = &cobra.Command{
    Use:    "shell-init",
    Hidden: true,  // User doesn't need to see this in help
    Short:  "Output shell wrapper function",
    RunE: func(cmd *cobra.Command, args []string) error {
        shellType, err := shell.DetectShell()
        if err != nil {
            return err
        }

        wrapper, err := shell.GetWrapper(shellType)
        if err != nil {
            return err
        }

        // Output to stdout for eval
        fmt.Print(wrapper)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(shellInitCmd)
}
```

### Dynamic Worktree Completion (Go)

```go
// Source: Cobra ValidArgsFunction documentation
package cmd

import (
    "path/filepath"

    "github.com/ahmedelarabyy/wt/internal/git"
    "github.com/spf13/cobra"
)

func worktreeNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    // Don't complete if we already have enough args
    if len(args) >= 1 {
        return nil, cobra.ShellCompDirectiveNoFileComp
    }

    // Query git worktree list
    worktrees, err := git.ListWorktrees()
    if err != nil {
        return nil, cobra.ShellCompDirectiveError
    }

    // Extract worktree names (basename or suffix after repo name)
    names := make([]string, 0, len(worktrees))
    for _, wt := range worktrees {
        name := extractWorktreeName(wt.Path)
        if name != "" {
            names = append(names, name)
        }
    }

    return names, cobra.ShellCompDirectiveNoFileComp
}

func extractWorktreeName(path string) string {
    // For "wt-feature" or "wt/feature", return "feature"
    basename := filepath.Base(path)

    // If this is home worktree (no suffix), skip it
    // Actual logic depends on Phase 3 suffix resolution
    // For now, return basename
    return basename
}

var gotoCmd = &cobra.Command{
    Use:               "goto <worktree>",
    Short:             "Switch to a worktree",
    ValidArgsFunction: worktreeNameCompletion,
    Args:              cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        // Command implementation...
        return nil
    },
}
```

### Cobra Completion Command Setup

```go
// Source: Cobra automatic completion
// This is AUTOMATIC in Cobra - just documenting what it does

// Cobra adds this command automatically:
// $ wt completion bash
// $ wt completion zsh
// $ wt completion fish

// User installs like:
// wt completion bash > /etc/bash_completion.d/wt
// wt completion zsh > /usr/local/share/zsh/site-functions/_wt
// wt completion fish > ~/.config/fish/completions/wt.fish

// No code needed! Cobra handles everything if ValidArgsFunction is set.
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Separate .sh/.bash/.zsh files | go:embed bundled scripts | Go 1.16 (2021) | Single binary distribution; no missing files |
| Static completions via _arguments array | ValidArgsFunction dynamic queries | Cobra 1.13+ (2020) | Always accurate completions; no stale cache |
| Manual completion script writing | Cobra auto-generation | Cobra 1.0+ (2017) | Cross-shell compatibility automatic; less code |
| Subshell cd (can't work) | Shell wrapper function | Always (fundamental limitation) | Only wrapper pattern works; child processes can't change parent |
| `[[` bash test operator | `[` POSIX test | N/A (POSIX compliance) | Works on bash 3.2 (macOS); more portable |
| Bundled completion + wrapper | Separate completion and shell-init | Modern pattern (2020+) | User choice; opt-in completions; cleaner separation |

**Deprecated/outdated:**
- **go-bindata for embedding:** Replaced by native `go:embed` in Go 1.16+
- **bash-completion v1 format:** Most systems now use v2 (Cobra supports both automatically)
- **Completion caching strategies:** Dynamic queries fast enough; caching adds complexity for minimal gain
- **Complex template systems for wrappers:** Static embedded strings sufficient; wrappers are simple
- **Shell version detection in runtime:** Just support bash 3.2+ baseline; modern shells are backward compatible

## Open Questions

Things that couldn't be fully resolved:

1. **Exact worktree info format on successful cd**
   - What we know: User decided "show worktree info (name + branch + path)" on successful cd
   - What's unclear: Exact format - one line or multiple? Separator style? Color?
   - Recommendation: Follow existing pattern from Phase 5 home/goto commands which print `"Switched to %s (branch: %s, %s)\n"`. For Phase 6, add path as third line: `"Path: %s\n"`. Or keep single line: `"Switched to %s (branch: %s, %s) at %s\n"`. Planner should choose based on existing Phase 5 confirmation format.

2. **Route all commands through wrapper vs only cd commands**
   - What we know: User marked as "Claude's discretion", leaning toward "all through wrapper"
   - What's unclear: merge/rebase don't need cd, but routing through wrapper adds consistency
   - Recommendation: Route all commands through wrapper for simplicity. Wrapper detects when binary outputs path (non-empty stdout) and only then runs cd. merge/rebase output nothing to stdout, so wrapper just returns their exit code. Simpler mental model, single code path, matches zoxide pattern.

3. **Minimum bash version explicit documentation**
   - What we know: User leaning toward bash 3.2+ for macOS, wrapper is simple
   - What's unclear: Should we test and document minimum versions for all shells?
   - Recommendation: Document bash 3.2+ (macOS Catalina default), zsh 5.0+ (any modern system), fish 3.0+ (maintains 2 years of releases). Test wrapper on macOS bash 3.2 explicitly. These versions are old enough that 99%+ users have them.

## Sources

### Primary (HIGH confidence)

- [Cobra v1.10.2 documentation](https://cobra.dev/docs/how-to-guides/shell-completion/) - Shell completion guide
- [Cobra package reference](https://pkg.go.dev/github.com/spf13/cobra) - ValidArgsFunction, ShellCompDirective API
- [Go embed package](https://pkg.go.dev/embed) - Official embedding documentation
- [Fish for bash users](https://fishshell.com/docs/current/fish_for_bash_users.html) - Fish syntax differences
- [How zoxide works](https://zoxide.org/en/blog/how-zoxide-works-en/) - Shell wrapper pattern explanation
- [zoxide GitHub](https://github.com/ajeetdsouza/zoxide) - Shell integration reference implementation

### Secondary (MEDIUM confidence)

- [Shell completions with Cobra blog](https://blog.chmouel.com/posts/cobra-completions/) - Practical completion examples
- [Cobra completions in Go blog](https://blog.devgenius.io/shell-completion-with-cobra-and-go-c8368074d8f7) - Implementation guide
- [Using Go Embed for templates](https://andrew-mccall.com/blog/2025/01/using-go-embed-package-for-template-rendering/) - Embedding patterns
- [Bash bypass alias](https://www.cyberciti.biz/faq/bash-bypass-alias-command-on-linux-macos-unix/) - Command builtin usage
- [Wrapping shell commands](https://kevinjalbert.com/wrapping-shell-commands-and-keep-the-original-name/) - Wrapper function patterns
- [Upgrading Bash on macOS](https://itnext.io/upgrading-bash-on-macos-7138bd1066ba) - Bash 3.2 compatibility context

### Tertiary (LOW confidence)

- [Shell wrapper examples](https://www.cyberciti.biz/tips/unix-linux-bash-shell-script-wrapper-examples.html) - General patterns
- [Advanced Bash scripting guide](https://www.linuxtopia.org/online_books/advanced_bash_scripting_guide/wrapper.html) - Wrapper concepts
- Web search results about shell integration patterns (2026)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Cobra is well-documented, go:embed is official Go feature, zoxide pattern is proven
- Architecture: HIGH - Patterns verified from official docs and reference implementations
- Pitfalls: MEDIUM-HIGH - Bash 3.2 limitations documented, fish differences from official docs, wrapper recursion is common knowledge, some pitfalls from experience

**Research date:** 2026-02-07
**Valid until:** 60 days (stable domain; Cobra updates infrequently; shell syntax unchanging)

**Key findings:**
- Cobra v1.10.2 handles all completion generation automatically - no manual script writing
- go:embed is the standard for bundling shell scripts (Go 1.16+)
- Dynamic ValidArgsFunction queries are fast enough; don't cache
- Bash 3.2 compatibility requires avoiding modern features (no `[[` with `==`, no associative arrays)
- Fish syntax is fundamentally different from bash; maintain separate templates
- Shell wrapper must use `command` builtin to avoid recursion
- All three shells (bash/zsh/fish) supported in single phase is feasible
