# Phase 3: Core Go Binary Foundation - Research

**Researched:** 2026-02-07
**Domain:** Go CLI application development with git worktree integration
**Confidence:** HIGH

## Summary

Phase 3 establishes the foundation of the Go-based `wt` tool by implementing commands that don't require shell integration: `wt --version`, `wt list`, `wt init`, and `wt delete`. The research reveals that the Go ecosystem has mature, well-established tooling for CLI development, with Cobra being the industry standard for command-line interfaces and standard patterns for interacting with git worktrees via command execution.

The key architectural decision is to **execute native git commands** rather than use pure Go git libraries (like go-git), as go-git v5 has limited worktree management capabilities—it doesn't support listing multiple worktrees or deleting them, which are core requirements for this phase. The mature approach is using os/exec to run git commands, with libraries like go-git-cmd-wrapper providing type-safe wrappers.

For terminal output, the fatih/color library provides automatic TTY detection and color management, aligning perfectly with the user's requirement for color auto-detection (color when stdout is a terminal, plain when piped).

**Primary recommendation:** Use Cobra (v1.10.2+) for CLI structure, execute git commands via os/exec (optionally wrapped with go-git-cmd-wrapper v2.9.1+), and use fatih/color (v1.18.0+) for terminal output with automatic color detection.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**List output format:**
- Default columns: name, branch, dirty/clean status indicator
- Path shown only with a flag (e.g., `-a` or `--all`)
- Current worktree marked with asterisk prefix (`*`) — like `git branch`
- Color auto-detected: color when stdout is a terminal, plain when piped
- Empty state: silent (no output, exit 0) — script-friendly

**Delete behavior:**
- No confirmation for clean worktrees — deletes immediately
- Dirty worktrees prompt for confirmation: "Worktree 'foo' has uncommitted changes. Delete? [y/N]"
- `--force` skips all confirmation, even for dirty worktrees
- Branch NOT deleted by default — `--branch` flag to also delete the branch
- Cannot delete current worktree — error: "can't delete current worktree"
- Single worktree per invocation (no multi-delete)
- Silent on success — no output on successful delete

**Error & output style:**
- Terse and direct — git-style messages
- Errors prefixed with `error:` (e.g., `error: worktree 'foo' not found`)
- Errors to stderr, data to stdout
- Simple commands silent on success (init, delete)
- Multi-step commands report each action (copy, symlink, run) — applies to Phase 5's `wt new`
- No `--quiet` flag — commands are already minimal

**Init defaults:**
- Template contains commented-out examples showing copy/symlink/run syntax
- Config format uses abstract actions: `copy`, `symlink` (Go handles cross-platform), plus `run` as escape hatch for custom commands
- Error if .wtconfig already exists — no overwrite (no --force)
- Created in current directory (not necessarily repo root)
- Requires being inside a git repo — error if not
- No auto-detection of project files — just the template

### Claude's Discretion

- Dirty/clean status indicator style (symbol choice)
- Exact list column alignment and spacing
- Commented example content in .wtconfig template
- Go project structure (module layout, package organization)
- Build system and CI tooling choices

### Deferred Ideas (OUT OF SCOPE)

- Config format details and parsing logic — Phase 4 (Configuration System)
- Multi-step action reporting for `wt new` — Phase 5 (Directory-Changing Commands)
- Cross-platform shell commands via `run` action — Phase 4/5

</user_constraints>

## Standard Stack

The established libraries/tools for Go CLI development with git integration:

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| [spf13/cobra](https://github.com/spf13/cobra) | v1.10.2+ | CLI framework and command structure | Industry standard used by Kubernetes, Docker, Hugo. Mature ecosystem with POSIX-compliant flags, auto-help, shell completion. |
| os/exec | stdlib | Execute git commands | Standard library approach for running external commands. Reliable, well-tested, no dependencies. |
| [fatih/color](https://github.com/fatih/color) | v1.18.0+ | Terminal color output with TTY detection | De facto standard for colored terminal output. Automatic TTY detection via isatty, respects NO_COLOR. |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| [ldez/go-git-cmd-wrapper](https://pkg.go.dev/github.com/ldez/go-git-cmd-wrapper/v2) | v2.9.1+ | Type-safe git command builder | Optional: provides builder pattern for constructing git commands. Better than string concatenation for complex commands. |
| [mattn/go-isatty](https://github.com/mattn/go-isatty) | latest | TTY detection | Usually transitive via fatih/color. Direct use if custom TTY detection needed. |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| os/exec for git | [go-git/go-git](https://pkg.go.dev/github.com/go-git/go-git/v5) v5.16.4 | Pure Go git implementation. **Limitation:** Doesn't support multiple worktrees (only one worktree per repository), can't list or delete worktrees. Missing critical features for this project. |
| spf13/cobra | [urfave/cli](https://github.com/urfave/cli) | Simpler but less feature-rich. Cobra's ecosystem and conventions are more established. |
| fatih/color | [gookit/color](https://pkg.go.dev/github.com/gookit/color) | More features but heavier. fatih/color is sufficient and lighter. |

**Installation:**
```bash
go get -u github.com/spf13/cobra@latest
go get -u github.com/fatih/color@latest
# Optional: for type-safe git command building
go get -u github.com/ldez/go-git-cmd-wrapper/v2@latest
```

## Architecture Patterns

### Recommended Project Structure

For a CLI tool of moderate complexity (single binary, multiple commands):

```
wt/
├── cmd/                    # Command implementations (one file per command)
│   ├── root.go            # Root command and shared setup
│   ├── version.go         # --version flag handler
│   ├── list.go            # wt list
│   ├── init.go            # wt init
│   └── delete.go          # wt delete
├── internal/              # Private application code
│   ├── git/               # Git operations (worktree list, status, delete)
│   ├── worktree/          # Worktree name resolution, path mapping
│   ├── config/            # .wtconfig template (Phase 4 will expand this)
│   └── output/            # Terminal output formatting (colors, tables)
├── main.go                # Entry point (minimal, just calls cmd.Execute())
├── go.mod
├── go.sum
└── README.md
```

**Rationale:**
- `/cmd` for command definitions keeps CLI logic separate and testable
- `/internal` prevents external imports and signals private implementation
- No `/pkg` needed—this is a single binary, not a library
- Flat structure inside /internal for now—can add subdirectories as complexity grows

**Source:** [golang-standards/project-layout](https://github.com/golang-standards/project-layout), [Go project structure best practices](https://www.alexedwards.net/blog/11-tips-for-structuring-your-go-projects)

### Pattern 1: Cobra Command Structure with Error Handling

**What:** Use `RunE` instead of `Run` for commands that can fail. Return errors to Cobra; it handles exit codes and error display.

**When to use:** All commands except trivial ones (like `--version`)

**Example:**
```go
// Source: https://www.jetbrains.com/guide/go/tutorials/cli-apps-go-cobra/error_handling/
var deleteCmd = &cobra.Command{
    Use:   "delete <worktree>",
    Short: "Remove a worktree",
    Args:  cobra.ExactArgs(1),
    // Use RunE to return errors
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]

        // Business logic returns errors
        if err := worktree.Delete(name); err != nil {
            return err  // Cobra handles exit code
        }

        // Silent on success (user requirement)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(deleteCmd)
    // Add flags
    deleteCmd.Flags().BoolP("force", "f", false, "skip confirmation")
    deleteCmd.Flags().Bool("branch", false, "also delete the branch")
}
```

**Key points:**
- `RunE` returns `error`, Cobra sets exit code 1 on non-nil error
- `Args: cobra.ExactArgs(1)` validates argument count automatically
- Errors propagate up; main.go doesn't call os.Exit
- Use `cmd.Flags()` for command-specific flags
- `cmd.SilenceUsage = true` prevents usage on runtime errors (do this in root command)

**Source:** [Cobra error handling guide](https://www.jetbrains.com/guide/go/tutorials/cli-apps-go-cobra/error_handling/), [Cobra documentation](https://cobra.dev/)

### Pattern 2: Git Command Execution with Error Handling

**What:** Execute git commands via os/exec, capture output and errors

**When to use:** All git operations (list worktrees, check status, delete, etc.)

**Example:**
```go
// Source: Go stdlib and community practices
func ListWorktrees() ([]Worktree, error) {
    cmd := exec.Command("git", "worktree", "list", "--porcelain")

    // Capture stdout
    output, err := cmd.Output()
    if err != nil {
        // Check for exit error to get stderr
        if exitErr, ok := err.(*exec.ExitError); ok {
            return nil, fmt.Errorf("git worktree list failed: %s", exitErr.Stderr)
        }
        return nil, fmt.Errorf("failed to execute git: %w", err)
    }

    // Parse porcelain output
    worktrees, err := parseWorktreeList(string(output))
    if err != nil {
        return nil, fmt.Errorf("failed to parse git output: %w", err)
    }

    return worktrees, nil
}

// With go-git-cmd-wrapper (type-safe alternative):
import (
    "github.com/ldez/go-git-cmd-wrapper/v2/git"
    "github.com/ldez/go-git-cmd-wrapper/v2/worktree"
)

func ListWorktreesTypeSafe() (string, error) {
    // Returns (output, error)
    return git.Worktree(worktree.List, worktree.Porcelain)
}
```

**Key points:**
- Use `--porcelain` flag for machine-readable git output
- `cmd.Output()` captures stdout; use `cmd.CombinedOutput()` if you need stderr too
- Wrap errors with context (`fmt.Errorf` with `%w` for error chains)
- Parse porcelain format carefully—it's line-based with key-value structure

**Source:** [go-git-cmd-wrapper docs](https://pkg.go.dev/github.com/ldez/go-git-cmd-wrapper/v2), Go stdlib exec package

### Pattern 3: Terminal Color Auto-Detection

**What:** Automatically enable color when stdout is a TTY, disable when piped

**When to use:** All user-facing output that might use color

**Example:**
```go
// Source: https://pkg.go.dev/github.com/fatih/color
import "github.com/fatih/color"

func PrintWorktreeList(worktrees []Worktree, current string) {
    // color.NoColor is automatically set based on:
    // - NO_COLOR env var
    // - TERM=dumb
    // - isatty check (stdout is not a TTY)
    // No manual setup needed!

    cyan := color.New(color.FgCyan)
    bold := color.New(color.Bold)

    for _, wt := range worktrees {
        marker := " "
        if wt.Path == current {
            marker = "*"
            // Print with bold for current worktree
            bold.Printf("%s %-30s %s\n", marker, wt.Name, wt.Branch)
        } else {
            // Normal print (respects NoColor automatically)
            cyan.Printf("%s %-30s %s\n", marker, wt.Name, wt.Branch)
        }
    }
}

// To force disable (for testing):
func init() {
    color.NoColor = true  // Disable globally
}
```

**Key points:**
- `color.NoColor` is auto-set—no manual isatty checks needed
- Each `color.New()` creates a reusable color function
- Colors automatically disabled when piping output
- Respects `NO_COLOR` environment variable (standard convention)

**Source:** [fatih/color documentation](https://pkg.go.dev/github.com/fatih/color)

### Pattern 4: Worktree Name Resolution (from wt.zsh)

**What:** Match user input against worktree directory basenames using suffix matching

**When to use:** Commands that accept worktree name argument (delete, goto in later phases)

**Example (adapted from wt.zsh _wt_resolve_path):**
```go
// Source: Existing wt.zsh implementation pattern
func ResolveWorktreeName(name string) (*Worktree, error) {
    worktrees, err := ListWorktrees()
    if err != nil {
        return nil, err
    }

    for _, wt := range worktrees {
        // Get directory basename
        dir := filepath.Base(wt.Path)

        // Match: exact name OR suffix after repo prefix
        // e.g. "my-repo-feature" matches input "feature"
        if dir == name || strings.HasSuffix(dir, "-"+name) {
            return &wt, nil
        }
    }

    return nil, fmt.Errorf("worktree '%s' not found", name)
}
```

**Key points:**
- Suffix matching enables short names: `wt delete feature` instead of `wt delete my-repo-feature`
- Exact match takes precedence (check `dir == name` first)
- Single match only—multiple matches should error (user requirement: INFRA-03)

**Source:** Existing wt.zsh implementation (lines 448-460)

### Anti-Patterns to Avoid

- **Calling os.Exit directly in command handlers**: Return errors instead. Let Cobra/main handle exit codes. Calling os.Exit makes testing impossible.
  - **Why it's bad:** [Testing os.Exit requires subprocess patterns](https://willsena.dev/golang-how-to-test-code-that-exits-or-crashes/). Error returns enable normal unit tests.
  - **What to do instead:** Use `RunE` and return errors. Set up proper error handling in root command.

- **Using panic for normal errors**: Panics are for programmer errors, not user/runtime errors like "worktree not found"
  - **Why it's bad:** Panics create ugly stack traces and can't be tested easily
  - **What to do instead:** Return typed errors with context

- **Writing to stdout and stderr inconsistently**: Data to stdout, errors to stderr
  - **Why it's bad:** Breaks shell pipelines and error redirection
  - **What to do instead:** Use `cmd.OutOrStdout()` and `cmd.ErrOrStderr()` in Cobra commands for testability

- **String concatenation for command building**: Fragile and error-prone
  - **Why it's bad:** Easy to introduce shell injection bugs, hard to test individual options
  - **What to do instead:** Use os/exec with argument slices OR go-git-cmd-wrapper's builder pattern

**Sources:** [Go os.Exit best practices](https://thelinuxcode.com/os-exit-golang-2/), [Cobra error handling](https://github.com/spf13/cobra/issues/914)

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| CLI argument parsing | Custom flag parser | spf13/cobra + pflag | POSIX compliance, subcommands, help generation, shell completion are complex. Cobra handles all edge cases. |
| TTY detection for colors | Custom isatty check | fatih/color (includes mattn/go-isatty) | Cross-platform TTY detection has platform-specific syscalls (Windows vs Unix). Library handles all platforms. |
| Git worktree operations | Pure Go git implementation | Execute native git commands via os/exec | go-git v5 lacks worktree management features. Native git is fully-featured and installed on all systems. |
| Terminal color codes | Manual ANSI escape sequences | fatih/color | ANSI codes vary by terminal, NO_COLOR convention, graceful degradation. Library handles compatibility. |
| Git porcelain parsing | Regex on human output | `git --porcelain` + structured parsing | Porcelain format is stable across git versions. Human output changes. |

**Key insight:** The Go ecosystem is mature for CLI development. Cobra is not just a library—it's a convention used across major projects. Using standard tools means familiar patterns for Go developers and better maintainability.

**Sources:** [Cobra usage in major projects](https://github.com/spf13/cobra), [go-git limitations](https://pkg.go.dev/github.com/go-git/go-git/v5)

## Common Pitfalls

### Pitfall 1: Using go-git for Worktree Management

**What goes wrong:** go-git v5 (latest stable) only supports a single worktree per repository object. You can't list multiple worktrees or delete them programmatically.

**Why it happens:** go-git is designed for porcelain operations (clone, commit, push) not plumbing (multi-worktree management). The library's architecture assumes one worktree per repo.

**How to avoid:** Use native git commands via os/exec. Git's worktree commands (`git worktree list --porcelain`, `git worktree remove`) are stable and well-documented.

**Warning signs:**
- Trying to call `repository.Worktrees()` (doesn't exist)
- Finding only `repository.Worktree()` which returns single worktree
- No `Worktree.Delete()` or `Worktree.List()` methods in go-git docs

**Source:** [go-git v5 documentation](https://pkg.go.dev/github.com/go-git/go-git/v5)—only `Repository.Worktree()` (singular) exists

### Pitfall 2: Calling os.Exit in Command Handlers

**What goes wrong:** Tests can't run because os.Exit terminates the test process. Code becomes untestable without subprocess hacks.

**Why it happens:** Coming from scripting backgrounds (bash, zsh) where `exit 1` is idiomatic. In Go, os.Exit skips deferred cleanup and makes testing require subprocess spawning.

**How to avoid:** Use Cobra's `RunE` instead of `Run`. Return errors, let Cobra handle exit codes. Only call os.Exit once in main.go:

```go
// main.go
func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)  // Only place os.Exit should appear
    }
}
```

**Warning signs:**
- `os.Exit` in command handlers
- Tests that use `exec.Command` to run own binary
- Deferred cleanup (closing files, temp dirs) not running

**Sources:** [Go testing with os.Exit](https://willsena.dev/golang-how-to-test-code-that-exits-or-crashes/), [Cobra error handling patterns](https://github.com/spf13/cobra/issues/914)

### Pitfall 3: Stdout/Stderr Confusion

**What goes wrong:** Errors mixed with data in stdout break shell pipelines. `wt list | grep foo` includes error messages in grep input.

**Why it happens:** Not explicitly routing output. Default `fmt.Println()` goes to stdout, but errors should go to stderr.

**How to avoid:** In Cobra commands, use:
- `cmd.OutOrStdout()` for data output (list results, etc.)
- `cmd.ErrOrStderr()` for errors
- `fmt.Fprintf(cmd.ErrOrStderr(), "error: %s\n", msg)` for error messages

This also makes testing easier—you can capture stdout and stderr separately.

**Warning signs:**
- Using `fmt.Println()` directly in command handlers
- Errors appearing when piping command output
- Inability to separate data from errors in tests

**Source:** [Cobra command patterns](https://cobra.dev/), Go stdlib best practices

### Pitfall 4: Not Using --porcelain for Git Output

**What goes wrong:** Parsing human-readable git output breaks when git updates output format or when user has custom git config (colors, column layout).

**Why it happens:** Human output is prettier and easier to read during development. Temptation to parse `git worktree list` without `--porcelain`.

**How to avoid:** Always use `--porcelain` flags for machine-readable output:
- `git worktree list --porcelain` (stable line-based format)
- `git status --porcelain` (single letter codes, easy to parse)

The porcelain format is stable across git versions and immune to user configuration.

**Warning signs:**
- Parsing `git worktree list` output with regex on column positions
- Splitting by whitespace (breaks with spaces in paths)
- Tests fail on different machines due to git config

**Source:** [Git documentation on porcelain commands](https://git-scm.com/docs/git-worktree), [go-git-cmd-wrapper examples](https://pkg.go.dev/github.com/ldez/go-git-cmd-wrapper/v2)

### Pitfall 5: Ignoring Cross-Platform Path Handling

**What goes wrong:** Hardcoded `/` path separators break on Windows. Operations like `strings.Split(path, "/")` fail.

**Why it happens:** Development on Unix-like systems. Path separator assumptions baked into string operations.

**How to avoid:**
- Use `filepath` package, not `path` (filepath is OS-aware)
- `filepath.Join()` for building paths
- `filepath.Base()` for directory name (not `strings.Split`)
- `filepath.Separator` if you must check separators

**Warning signs:**
- Using `path` package (for URLs) instead of `filepath` (for filesystem)
- Hardcoded `/` or `\` in path operations
- String manipulation on paths

**Source:** Go stdlib documentation, cross-platform best practices

## Code Examples

Verified patterns from official sources and existing wt.zsh:

### Basic Cobra Command Setup

```go
// cmd/root.go
// Source: https://github.com/spf13/cobra user guide
package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "wt",
    Short: "Git worktree manager",
    Long:  `wt is a fast, intuitive git worktree manager`,
    // Don't show usage on runtime errors
    SilenceUsage: true,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    // Global flags
    rootCmd.PersistentFlags().BoolP("help", "h", false, "show help")
}
```

```go
// main.go
package main

import (
    "os"
    "github.com/yourusername/wt/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Git Worktree List Implementation

```go
// internal/git/worktree.go
// Based on: wt.zsh _wt_list function and git porcelain format
package git

import (
    "fmt"
    "os/exec"
    "strings"
)

type Worktree struct {
    Path   string
    Branch string
    Locked bool
    Prunable bool
}

func ListWorktrees() ([]Worktree, error) {
    cmd := exec.Command("git", "worktree", "list", "--porcelain")
    output, err := cmd.Output()
    if err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            return nil, fmt.Errorf("git worktree list: %s", exitErr.Stderr)
        }
        return nil, err
    }

    return parseWorktreePorcelain(string(output)), nil
}

func parseWorktreePorcelain(output string) []Worktree {
    var worktrees []Worktree
    var current Worktree

    lines := strings.Split(output, "\n")
    for _, line := range lines {
        if line == "" {
            // Empty line separates worktrees
            if current.Path != "" {
                worktrees = append(worktrees, current)
                current = Worktree{}
            }
            continue
        }

        if strings.HasPrefix(line, "worktree ") {
            current.Path = strings.TrimPrefix(line, "worktree ")
        } else if strings.HasPrefix(line, "branch ") {
            current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
        } else if strings.HasPrefix(line, "locked") {
            current.Locked = true
        } else if strings.HasPrefix(line, "prunable") {
            current.Prunable = true
        }
    }

    // Last entry (no trailing newline)
    if current.Path != "" {
        worktrees = append(worktrees, current)
    }

    return worktrees
}
```

### Dirty/Clean Status Check

```go
// internal/git/status.go
package git

import (
    "os/exec"
    "strings"
)

// IsWorktreeDirty returns true if worktree has uncommitted changes
func IsWorktreeDirty(path string) (bool, error) {
    cmd := exec.Command("git", "-C", path, "status", "--porcelain")
    output, err := cmd.Output()
    if err != nil {
        return false, err
    }

    // Porcelain output: empty if clean, non-empty if dirty
    return strings.TrimSpace(string(output)) != "", nil
}
```

### Interactive Confirmation Prompt

```go
// internal/output/prompt.go
package output

import (
    "bufio"
    "fmt"
    "io"
    "strings"
)

// Confirm prompts user for yes/no confirmation. Default is no.
func Confirm(prompt string, reader io.Reader) (bool, error) {
    fmt.Printf("%s [y/N] ", prompt)

    scanner := bufio.NewScanner(reader)
    if !scanner.Scan() {
        return false, scanner.Err()
    }

    response := strings.ToLower(strings.TrimSpace(scanner.Text()))
    return response == "y" || response == "yes", nil
}
```

### List Command Implementation

```go
// cmd/list.go
// Based on: wt.zsh _wt_list function
package cmd

import (
    "fmt"
    "path/filepath"

    "github.com/fatih/color"
    "github.com/spf13/cobra"
    "github.com/yourusername/wt/internal/git"
)

var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List all worktrees",
    RunE: func(cmd *cobra.Command, args []string) error {
        worktrees, err := git.ListWorktrees()
        if err != nil {
            return fmt.Errorf("error: failed to list worktrees: %w", err)
        }

        // Empty state: silent (user requirement)
        if len(worktrees) == 0 {
            return nil
        }

        currentPath, _ := git.CurrentWorktree()

        // Colors auto-detected by fatih/color
        cyan := color.New(color.FgCyan)

        for _, wt := range worktrees {
            marker := " "
            if wt.Path == currentPath {
                marker = "*"
            }

            name := filepath.Base(wt.Path)

            // Check dirty/clean status
            dirty, _ := git.IsWorktreeDirty(wt.Path)
            status := "✓"  // clean
            if dirty {
                status = "✗"  // dirty (symbol choice is Claude's discretion)
            }

            // Color when TTY, plain when piped (automatic)
            cyan.Printf("%s %-30s %-20s %s\n", marker, name, wt.Branch, status)
        }

        return nil
    },
}

func init() {
    rootCmd.AddCommand(listCmd)
}
```

### Delete Command Implementation

```go
// cmd/delete.go
// Based on: user requirements from CONTEXT.md
package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/yourusername/wt/internal/git"
    "github.com/yourusername/wt/internal/output"
    "github.com/yourusername/wt/internal/worktree"
)

var (
    forceDelete  bool
    deleteBranch bool
)

var deleteCmd = &cobra.Command{
    Use:   "delete <worktree>",
    Short: "Remove a worktree",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]

        // Resolve name to worktree
        wt, err := worktree.Resolve(name)
        if err != nil {
            return fmt.Errorf("error: worktree '%s' not found", name)
        }

        // Cannot delete current worktree
        current, _ := git.CurrentWorktree()
        if wt.Path == current {
            return fmt.Errorf("error: can't delete current worktree")
        }

        // Check if dirty
        dirty, err := git.IsWorktreeDirty(wt.Path)
        if err != nil {
            return fmt.Errorf("error: failed to check worktree status: %w", err)
        }

        // Prompt for confirmation if dirty (unless --force)
        if dirty && !forceDelete {
            confirmed, err := output.Confirm(
                fmt.Sprintf("Worktree '%s' has uncommitted changes. Delete?", name),
                os.Stdin,
            )
            if err != nil {
                return err
            }
            if !confirmed {
                return nil  // Silent exit on no
            }
        }

        // Delete worktree
        if err := git.RemoveWorktree(wt.Path); err != nil {
            return fmt.Errorf("error: failed to remove worktree: %w", err)
        }

        // Delete branch if requested
        if deleteBranch && wt.Branch != "" {
            if err := git.DeleteBranch(wt.Branch); err != nil {
                return fmt.Errorf("error: failed to delete branch: %w", err)
            }
        }

        // Silent on success (user requirement)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(deleteCmd)
    deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "skip confirmation")
    deleteCmd.Flags().BoolVar(&deleteBranch, "branch", false, "also delete the branch")
}
```

### Init Command Implementation

```go
// cmd/init.go
package cmd

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
    "github.com/yourusername/wt/internal/config"
)

var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Create .wtconfig template",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Must be in git repo
        repoRoot, err := git.RepoRoot()
        if err != nil {
            return fmt.Errorf("error: not inside a git repository")
        }

        // Create in current directory (not repo root - user requirement)
        cwd, _ := os.Getwd()
        configPath := filepath.Join(cwd, ".wtconfig")

        // Error if exists
        if _, err := os.Stat(configPath); err == nil {
            return fmt.Errorf("error: .wtconfig already exists")
        }

        // Write template
        if err := config.WriteTemplate(configPath); err != nil {
            return fmt.Errorf("error: failed to create config: %w", err)
        }

        // Report success
        fmt.Printf("Created %s\n", configPath)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(initCmd)
}
```

```go
// internal/config/template.go
// Based on: wt.zsh _wt_init function
package config

import (
    "os"
)

const template = `# .wtconfig — files to copy or symlink into new worktrees
# Syntax: <action> <path>
# Actions: copy, symlink, run

# Node.js
# copy .env.local
# symlink node_modules
# run npm install

# Python
# copy .env
# symlink .venv
# run pip install -e .

# Rust
# symlink target
# run cargo build
`

func WriteTemplate(path string) error {
    return os.WriteFile(path, []byte(template), 0644)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| String-based CLI parsing | Cobra with typed flags | 2019+ | Type safety, auto-help, shell completion. Cobra v1.0 released 2019. |
| go-git for all git ops | Native git via os/exec for worktrees | Current necessity | go-git doesn't support multiple worktrees. Native git is feature-complete. |
| Manual ANSI codes | fatih/color with auto TTY detection | Mature (5+ years) | NO_COLOR convention, cross-platform, automatic pipe detection. |
| Separate command files | cobra-cli generator | 2020+ | Boilerplate generation, but hand-writing is still common for control. |
| GOPATH-based projects | Go modules (go.mod) | Go 1.11+ (2018), default 1.16+ (2021) | Reproducible builds, version locking, no GOPATH required. |

**Deprecated/outdated:**
- **GOPATH workspace mode**: Replaced by Go modules. All new projects use `go.mod`.
- **dep dependency manager**: Replaced by Go modules. Don't use dep for new projects.
- **go-git v4**: Superseded by v5. Use v5.16.4+ for current API.

**Current state (2026):**
- Cobra v1.10.2 is mature and stable. No major breaking changes expected.
- Go 1.21+ has integrated toolchain management. Use `go 1.21` or later in go.mod.
- fatih/color v1.18.0 supports all modern terminals and respects NO_COLOR.

**Sources:** [Cobra releases](https://github.com/spf13/cobra/releases), [Go modules documentation](https://go.dev/doc/modules/), [fatih/color](https://github.com/fatih/color)

## Testing Patterns

### Table-Driven Tests for Commands

```go
// cmd/list_test.go
// Source: https://go.dev/wiki/TableDrivenTests
package cmd

import (
    "bytes"
    "testing"

    "github.com/spf13/cobra"
)

func TestListCommand(t *testing.T) {
    tests := []struct {
        name           string
        setupMock      func()
        expectedOutput string
        expectError    bool
    }{
        {
            name: "empty list",
            setupMock: func() {
                // Mock git to return no worktrees
            },
            expectedOutput: "",  // Silent on empty
            expectError:    false,
        },
        {
            name: "single worktree",
            setupMock: func() {
                // Mock git to return one worktree
            },
            expectedOutput: "* main-repo    main   ✓\n",
            expectError:    false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            tt.setupMock()

            // Capture output
            buf := new(bytes.Buffer)
            cmd := listCmd
            cmd.SetOut(buf)
            cmd.SetErr(buf)

            // Execute
            err := cmd.Execute()

            // Assert
            if tt.expectError && err == nil {
                t.Error("expected error, got nil")
            }
            if !tt.expectError && err != nil {
                t.Errorf("unexpected error: %v", err)
            }

            output := buf.String()
            if output != tt.expectedOutput {
                t.Errorf("expected %q, got %q", tt.expectedOutput, output)
            }
        })
    }
}
```

**Source:** [Table-driven tests in Go](https://medium.com/@mojimich2015/table-driven-tests-in-go-a-practical-guide-8135dcbc27ca), [Cobra testing patterns](https://gianarb.it/blog/golang-mockmania-cli-command-with-cobra)

## Build and Distribution

### GoReleaser for Cross-Platform Builds

**What:** Automated build system for creating binaries for multiple platforms

**Basic setup:**
```yaml
# .goreleaser.yml
project_name: wt

before:
  hooks:
    - go mod download
    - go test ./...

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{- .Version }}_
      {{- .Os }}_
      {{- .Arch }}
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
```

**Install:** `brew install goreleaser/tap/goreleaser`

**Local build test:** `goreleaser build --snapshot --clean`

**CI/CD (GitHub Actions):**
```yaml
name: Release
on:
  push:
    tags:
      - 'v*'
jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Source:** [GoReleaser documentation](https://goreleaser.com/), [How to publish Go binaries with GoReleaser](https://www.kosli.com/blog/how-to-publish-your-golang-binaries-with-goreleaser/)

## Open Questions

Things that couldn't be fully resolved:

1. **Dirty/clean status indicator symbol**
   - What we know: User wants a visual indicator, chose between symbols like ✓/✗, ●/○, or words
   - What's unclear: Exact symbol preference—this is marked as Claude's discretion
   - Recommendation: Use ✓ (U+2713) for clean, ✗ (U+2717) for dirty. UTF-8 widely supported. Fallback: words "clean"/"dirty" if terminal encoding issues.

2. **Worktree name ambiguity handling**
   - What we know: Suffix matching enables short names (INFRA-03 requirement)
   - What's unclear: What if two worktrees match? "my-repo-feat" and "other-repo-feat" both match "feat"
   - Recommendation: Phase 3 implements single-match requirement—error on ambiguous input. Phase 6 (Error Handling) can enhance with suggestions.

3. **Version string embedding**
   - What we know: Need `wt --version` to show version
   - What's unclear: Version source—hardcoded, git tag, or build-time injection?
   - Recommendation: Use ldflags injection via GoReleaser: `-X main.version={{.Version}}`. Fallback: "dev" for local builds.

## Sources

### Primary (HIGH confidence)

- [spf13/cobra GitHub](https://github.com/spf13/cobra) - v1.10.2 CLI framework
- [Cobra pkg.go.dev](https://pkg.go.dev/github.com/spf13/cobra) - API documentation
- [fatih/color GitHub](https://github.com/fatih/color) - v1.18.0 color library
- [fatih/color pkg.go.dev](https://pkg.go.dev/github.com/fatih/color) - Color API and auto-detection
- [go-git-cmd-wrapper pkg.go.dev](https://pkg.go.dev/github.com/ldez/go-git-cmd-wrapper/v2) - v2.9.1 git wrapper
- [go-git v5 pkg.go.dev](https://pkg.go.dev/github.com/go-git/go-git/v5) - v5.16.4 (worktree limitations verified)
- [Go official module documentation](https://go.dev/doc/modules/layout) - Module layout
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout) - Standard structure
- [Git worktree documentation](https://git-scm.com/docs/git-worktree) - Porcelain format
- Existing wt.zsh implementation - Worktree name resolution pattern

### Secondary (MEDIUM confidence)

- [JetBrains Cobra error handling guide](https://www.jetbrains.com/guide/go/tutorials/cli-apps-go-cobra/error_handling/) - RunE patterns
- [Alex Edwards Go project structure](https://www.alexedwards.net/blog/11-tips-for-structuring-your-go-projects) - 11 tips for structuring Go projects
- [GoReleaser documentation](https://goreleaser.com/) - Build automation
- [Table-driven tests Medium guide](https://medium.com/@mojimich2015/table-driven-tests-in-go-a-practical-guide-8135dcbc27ca) - Testing patterns (Jan 2026)
- [Cobra testing patterns](https://gianarb.it/blog/golang-mockmania-cli-command-with-cobra) - CLI testing with mocks

### Tertiary (LOW confidence - community practices)

- [Testing os.Exit in Go](https://willsena.dev/golang-how-to-test-code-that-exits-or-crashes/) - Subprocess pattern
- [Cobra GitHub issues on error handling](https://github.com/spf13/cobra/issues/914) - Community discussions
- [Go os.Exit usage guide](https://thelinuxcode.com/os-exit-golang-2/) - Best practices article

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Cobra, fatih/color, os/exec are well-documented with stable APIs
- Architecture: HIGH - Go project structure and Cobra patterns are well-established
- Pitfalls: MEDIUM - Based on community experience and documented limitations (e.g., go-git)
- Testing patterns: MEDIUM - Table-driven tests are standard, but Cobra testing examples vary

**Research date:** 2026-02-07
**Valid until:** 60 days (2026-04-07) - Stable ecosystem, slow-moving major versions
