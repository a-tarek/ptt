# Phase 4: Configuration System - Research

**Researched:** 2026-02-07
**Domain:** Plain text config parsing, file operations, command execution in Go
**Confidence:** HIGH

## Summary

Phase 4 implements a configuration system for worktree setup automation. The domain involves three main technical areas: (1) parsing plain text configuration files with a simple `action path` format, (2) performing file operations (copy, symlink) with proper parent directory creation, and (3) executing shell commands with real-time output streaming. The v1.0 implementation uses zsh with manual parsing and basic `cp`/`ln -s` operations; the v2.0 Go port requires robust error handling, upfront validation, and atomic rollback on failure.

The standard Go approach uses stdlib for most operations with one well-established third-party library (otiai10/copy) for recursive directory copying. Go's `os/exec` package handles command execution with output streaming via `cmd.Stdout` and `cmd.Stderr` assignment. Config parsing uses `bufio.Scanner` for line-by-line reading with `strings` package for trimming and prefix detection. The mutually exclusive config source model (file vs inline flags) maps cleanly to cobra's validation hooks.

**Primary recommendation:** Use Go stdlib for parsing and symlinks, otiai10/copy v1.14.1 for recursive copies, os/exec for command execution, and implement upfront validation + deferred rollback pattern with explicit error collection.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Config format:**
- Plain text format: `action path` per line, `#` comments, blank lines ignored
- Three actions: `copy`, `symlink`, `run` — all ported from v1.0
- No glob patterns — explicit paths only
- Files and directories both supported (recursive copy for directories)
- Parent directories created automatically when needed (like `mkdir -p`)
- `run` syntax: everything after `run ` is the command string, no quoting required
- `run` commands execute with the new worktree (target) as working directory
- `run` output streamed to user (stdout/stderr visible in real-time)
- Execution order: strict sequential as written in the file — no type grouping
- Config files live at repo root: `.wtconfig`, `.wtconfig-*`

**Config model (mutually exclusive sources):**
- One config source per invocation — no merging, no overriding:
  - `wt new <branch>` — uses `.wtconfig` (default)
  - `wt new <branch> --config <name>` — uses named config file
  - `wt new <branch> --copy X --symlink Y --run Z` — inline flags only, no config file read
- `--config` resolution: bare name → `.wtconfig-{name}` at repo root; contains `/` → treated as exact path
- Tab completion for `--config` lists available `.wtconfig-*` files (implementation in Phase 6)
- Inline flags can mix all three types freely: `--copy`, `--symlink`, `--run`
- Inline flag execution order: sequential in flag order as given on command line
- Duplicate flag paths = error (e.g., `--copy .env --symlink .env` is rejected)
- `--config nonexistent` = error, not fallback to no-config

**Init command updates:**
- `wt init` creates `.wtconfig` with template showing all three actions (copy, symlink, run) with examples
- `wt init --name foo` creates `.wtconfig-foo` with the same template
- Keep generic template — no project-type detection

**Failure & rollback:**
- Strict by default: missing source file for copy/symlink = failure
- Any action failure (copy, symlink, or run) aborts and rolls back
- Rollback = delete entire new worktree (git worktree remove + directory cleanup)
- Upfront validation: parse entire config and check all referenced files exist before executing any actions
- All validation errors reported at once (not fail-on-first)
- Run failures show both stderr output and exit code
- If rollback itself fails: warn with manual cleanup instructions, exit with error

**Config-free operation:**
- No .wtconfig + no flags = create worktree silently with zero setup actions
- No hint to run `wt init` — config is purely optional
- Success output: stream each action as it runs ("Copied .env", "Symlinked node_modules", "Running npm install..."), then show worktree path

### Claude's Discretion

- Exact validation error message formatting
- Internal config parsing implementation (structs, interfaces)
- How to detect mutually exclusive flag groups (config vs inline)
- Rollback implementation details (order of cleanup operations)

### Deferred Ideas (OUT OF SCOPE)

- **copyEnv action** — A config action that copies env files but prompts for variable overrides per worktree (e.g., `copyEnv .env.local --VITE_PORT`). Would enable different ports/credentials per worktree without manual editing. Significant feature with prompt handling, env file parsing, variable substitution — belongs in its own phase.

</user_constraints>

## Standard Stack

The established libraries/tools for this domain:

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| bufio.Scanner | stdlib | Line-by-line file parsing | Memory-efficient, built-in split functions, handles large files |
| strings | stdlib | String manipulation (HasPrefix, TrimSpace, etc.) | Core text processing, zero dependencies |
| os | stdlib | File operations, directory creation | Cross-platform file ops, os.MkdirAll for parent dirs |
| os/exec | stdlib | Shell command execution | Standard process spawning, I/O streaming support |
| path/filepath | stdlib | Path manipulation | Cross-platform path handling |
| github.com/otiai10/copy | v1.14.1 | Recursive directory copying | 1,264+ projects use it, handles edge cases (symlinks, permissions), MIT license |
| github.com/spf13/cobra | v1.10.2 | CLI framework (already in use) | Built-in flag validation, mutually exclusive flags support |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/fatih/color | v1.18.0 | Terminal colors (already in use) | Status output formatting |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| otiai10/copy | Manual os.CopyFS or custom recursion | Library handles symlinks, permissions, edge cases better. Manual solution would need extensive testing. |
| bufio.Scanner | os.ReadFile + strings.Split | Scanner is more memory-efficient for large files and provides cleaner line-by-line semantics. |
| cobra flag validation | Manual flag checking in RunE | Cobra's MarkFlagsMutuallyExclusive is clearer and generates better error messages. |

**Installation:**
```bash
go get github.com/otiai10/copy@v1.14.1
# cobra and color already in go.mod
```

## Architecture Patterns

### Recommended Project Structure

```
internal/
├── config/              # Config parsing and validation
│   ├── parser.go       # Parse .wtconfig files
│   ├── action.go       # Action type definitions
│   └── validator.go    # Upfront validation logic
├── setup/              # Worktree setup actions
│   ├── executor.go     # Execute actions with rollback
│   ├── copy.go         # Copy file/directory implementation
│   ├── symlink.go      # Symlink implementation
│   └── run.go          # Command execution with streaming
└── git/                # Git operations (already exists)
    └── worktree.go
```

### Pattern 1: Config Parsing with bufio.Scanner

**What:** Line-by-line parsing with whitespace/comment handling
**When to use:** Plain text config files with line-oriented syntax
**Example:**
```go
// Source: Go stdlib documentation + research findings
func parseConfig(path string) ([]Action, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var actions []Action
    scanner := bufio.NewScanner(file)
    lineNum := 0

    for scanner.Scan() {
        lineNum++
        line := strings.TrimSpace(scanner.Text())

        // Skip blank lines and comments
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        // Parse: "action path" or "run command with spaces"
        parts := strings.SplitN(line, " ", 2)
        if len(parts) < 2 {
            return nil, fmt.Errorf("line %d: invalid format", lineNum)
        }

        action := Action{
            Type: parts[0],
            Path: strings.TrimSpace(parts[1]),
        }
        actions = append(actions, action)
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return actions, nil
}
```

### Pattern 2: Upfront Validation with Collected Errors

**What:** Validate all actions before executing any, report all errors at once
**When to use:** Atomic operations requiring all-or-nothing execution
**Example:**
```go
// Source: Research best practices for validation patterns
func validateActions(srcRoot string, actions []Action) error {
    var errs []string

    for i, action := range actions {
        switch action.Type {
        case "copy", "symlink":
            // Check source file exists
            fullPath := filepath.Join(srcRoot, action.Path)
            if _, err := os.Stat(fullPath); err != nil {
                errs = append(errs, fmt.Sprintf("action %d: %s", i+1, err))
            }
        case "run":
            // Commands validated at execution time
            if action.Path == "" {
                errs = append(errs, fmt.Sprintf("action %d: empty command", i+1))
            }
        default:
            errs = append(errs, fmt.Sprintf("action %d: unknown action '%s'", i+1, action.Type))
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("validation failed:\n  %s", strings.Join(errs, "\n  "))
    }
    return nil
}
```

### Pattern 3: Deferred Rollback with Error Capture

**What:** Use defer for cleanup, capture both primary and rollback errors
**When to use:** Operations requiring atomic rollback on failure
**Example:**
```go
// Source: Official Go transaction pattern + research findings
func executeActions(srcRoot, targetRoot string, actions []Action) (err error) {
    // Track success state for rollback decision
    completed := false

    defer func() {
        if !completed && err != nil {
            // Rollback: remove entire worktree
            fmt.Fprintf(os.Stderr, "Error occurred, rolling back...\n")

            // git worktree remove
            removeCmd := exec.Command("git", "worktree", "remove", "--force", targetRoot)
            if removeErr := removeCmd.Run(); removeErr != nil {
                fmt.Fprintf(os.Stderr, "WARNING: rollback failed: %v\n", removeErr)
                fmt.Fprintf(os.Stderr, "Manual cleanup required: git worktree remove --force %s\n", targetRoot)
                // Return original error, not rollback error
                return
            }

            // Directory cleanup (if git remove left anything)
            if _, statErr := os.Stat(targetRoot); statErr == nil {
                if removeErr := os.RemoveAll(targetRoot); removeErr != nil {
                    fmt.Fprintf(os.Stderr, "WARNING: directory cleanup failed: %v\n", removeErr)
                    fmt.Fprintf(os.Stderr, "Manual cleanup required: rm -rf %s\n", targetRoot)
                }
            }
        }
    }()

    // Execute actions sequentially
    for _, action := range actions {
        if err = executeAction(srcRoot, targetRoot, action); err != nil {
            return err // triggers rollback via defer
        }
    }

    completed = true
    return nil
}
```

### Pattern 4: Command Execution with Real-Time Output Streaming

**What:** Stream stdout/stderr to user as command runs
**When to use:** Long-running commands where user needs progress feedback
**Example:**
```go
// Source: Go os/exec documentation + research findings
func executeRunAction(workDir, command string) error {
    cmd := exec.Command("sh", "-c", command)
    cmd.Dir = workDir

    // Stream output directly to user
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    fmt.Printf("Running: %s\n", command)

    if err := cmd.Run(); err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            return fmt.Errorf("command failed with exit code %d", exitErr.ExitCode())
        }
        return fmt.Errorf("command failed: %w", err)
    }

    return nil
}
```

### Pattern 5: Cobra Mutually Exclusive Flag Groups

**What:** Enforce that only one config source is used per invocation
**When to use:** CLI with mutually exclusive input modes
**Example:**
```go
// Source: Cobra documentation
var (
    configName   string
    copyPaths    []string
    symlinkPaths []string
    runCommands  []string
)

func init() {
    newCmd.Flags().StringVar(&configName, "config", "", "use named config variant")
    newCmd.Flags().StringSliceVar(&copyPaths, "copy", nil, "copy path (can be repeated)")
    newCmd.Flags().StringSliceVar(&symlinkPaths, "symlink", nil, "symlink path (can be repeated)")
    newCmd.Flags().StringSliceVar(&runCommands, "run", nil, "run command (can be repeated)")

    // NOTE: Cobra's MarkFlagsMutuallyExclusive doesn't support "one group vs another group"
    // Use PreRunE for custom validation instead
}

func validateConfigSource() error {
    hasConfigFlag := configName != ""
    hasInlineFlags := len(copyPaths) > 0 || len(symlinkPaths) > 0 || len(runCommands) > 0

    if hasConfigFlag && hasInlineFlags {
        return fmt.Errorf("cannot use --config with --copy/--symlink/--run flags")
    }

    return nil
}
```

### Pattern 6: Duplicate Detection with Map

**What:** Check for duplicate paths across inline flags
**When to use:** Validating user input for conflicting operations
**Example:**
```go
// Source: Go duplicate detection research
func checkDuplicatePaths(copyPaths, symlinkPaths []string) error {
    seen := make(map[string]string) // path -> action type

    for _, path := range copyPaths {
        if existing, ok := seen[path]; ok {
            return fmt.Errorf("duplicate path '%s' (specified as both --copy and --%s)", path, existing)
        }
        seen[path] = "copy"
    }

    for _, path := range symlinkPaths {
        if existing, ok := seen[path]; ok {
            return fmt.Errorf("duplicate path '%s' (specified as both --symlink and --%s)", path, existing)
        }
        seen[path] = "symlink"
    }

    return nil
}
```

### Anti-Patterns to Avoid

- **Fail-on-first validation:** Report all validation errors at once, not just the first one. Users shouldn't have to fix issues one at a time.
- **Silent skips on missing files:** Strict-by-default catches typos. Silent skip hides configuration bugs.
- **Buffered command output:** Streaming output to user provides progress feedback for long-running commands (npm install, cargo build, etc.)
- **Manual recursive copy:** Go's stdlib doesn't include recursive directory copy. Use otiai10/copy instead of rolling your own (handles symlinks, permissions, edge cases).
- **Path string splitting:** Use `strings.SplitN(line, " ", 2)` not `strings.Split(line, " ")` so that `run npm install --save` doesn't split into multiple parts.

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Recursive directory copy | Custom filepath.Walk + os.Copy loop | otiai10/copy | Handles symlinks (3 strategies), preserves permissions/timestamps, handles directory collisions, 1,264+ projects trust it |
| Config file parsing | Regex-based line parser | bufio.Scanner + strings package | Memory-efficient for large files, handles various line endings, cleaner error handling |
| Parent directory creation | Recursive mkdir loop | os.MkdirAll(path, 0755) | stdlib handles race conditions, idempotent (no error if exists), cross-platform |
| Mutually exclusive flags | Manual boolean checks | Cobra's validation in PreRunE | Clearer intent, consistent error messages, less boilerplate |
| Duplicate detection | Nested loops | Map-based tracking | O(n) vs O(n²), scales better, idiomatic Go |

**Key insight:** File operations have subtle edge cases (symlinks, permissions, race conditions). Using battle-tested libraries prevents bugs that only appear in specific environments (Windows vs Unix, NFS mounts, etc.).

## Common Pitfalls

### Pitfall 1: os.RemoveAll Memory Issues with Deep Trees

**What goes wrong:** Very deep directory trees (~2.5M subdirectories) can exhaust memory during os.RemoveAll, causing OOM killer to terminate the process without completing cleanup.

**Why it happens:** os.RemoveAll doesn't close file descriptors until after recursing to the bottom, keeping many files open simultaneously.

**How to avoid:**
- For Phase 4, this is unlikely (worktrees won't have millions of subdirectories)
- If rollback fails, warn user with manual cleanup instructions: `git worktree remove --force <path>`
- Don't silently swallow rollback errors

**Warning signs:** os.RemoveAll call hangs or process killed by OOM on large codebases with deep node_modules nesting.

**Sources:** [GitHub Issue #47390](https://github.com/golang/go/issues/47390), [GitHub Issue #20841](https://github.com/golang/go/issues/20841)

### Pitfall 2: Scanner Default Buffer Too Small for Long Lines

**What goes wrong:** bufio.Scanner has a default 64KB buffer. Lines longer than this cause "token too long" errors.

**Why it happens:** Config files might contain very long `run` commands with many arguments.

**How to avoid:**
- Use `scanner.Buffer(buf, maxCapacity)` to increase buffer if needed
- For Phase 4, unlikely to hit this (commands rarely exceed 64KB)
- If error occurs, improve error message: "line too long (max 64KB)"

**Warning signs:** Parse errors on config files with very long command strings.

**Sources:** [Go bufio documentation](https://pkg.go.dev/bufio)

### Pitfall 3: Symlink Behavior Differs on Windows

**What goes wrong:** Symlinks on Windows require Developer Mode or admin privileges in older versions. os.Symlink fails with permission errors for non-admin users.

**Why it happens:** Windows symlink creation is historically restricted for security reasons.

**How to avoid:**
- Document Windows requirement: enable Developer Mode (Windows 11/12 standard practice)
- Provide clear error message when os.Symlink fails with permission error
- Phase 4 spec says "cross-platform" but Windows symlink limitations are OS-level, not fixable in code

**Warning signs:** Symlink actions fail on Windows with permission denied errors for non-admin users.

**Sources:** [Cross-platform symlinks article](https://geedew.com/cross-platform-symlink/)

### Pitfall 4: Using os.Stat Instead of os.Lstat for Symlink Checks

**What goes wrong:** os.Stat follows symlinks to the target. If checking whether a path is a symlink, os.Stat returns info about the target, not the link itself.

**Why it happens:** os.Stat is for checking the *target* of a symlink, not the symlink itself.

**How to avoid:**
- Use os.Lstat when you need to know if something IS a symlink
- Use os.Stat when you need to validate the symlink's target exists
- For Phase 4 validation: use os.Stat (we want to verify source exists, don't care if it's a symlink)

**Warning signs:** Validation passes on broken symlinks that point to non-existent targets.

**Sources:** [Understanding os.Stat vs os.Lstat](https://dev.to/moseeh_52/understanding-osstat-vs-oslstat-in-go-file-and-symlink-handling-3p5d)

### Pitfall 5: Not Trimming Whitespace After Parsing

**What goes wrong:** Config line `copy .env  ` (trailing spaces) results in path `.env  ` which fails to match `.env` when checking existence.

**Why it happens:** User typos, editor trailing whitespace, copy-paste artifacts.

**How to avoid:**
- Always `strings.TrimSpace()` after parsing line
- Trim both the full line (to detect comments/blank lines) AND the path portion after splitting

**Warning signs:** Validation errors claiming files don't exist when they clearly do (path has invisible trailing whitespace).

**Sources:** Standard Go parsing best practices

### Pitfall 6: Forgetting to Check scanner.Err()

**What goes wrong:** After `scanner.Scan()` loop completes, errors reading the file (I/O errors, permission issues) are silently ignored unless you check `scanner.Err()`.

**Why it happens:** bufio.Scanner doesn't return errors from `Scan()`, it stores them and returns false. You must check explicitly.

**How to avoid:**
```go
for scanner.Scan() {
    // process line
}
if err := scanner.Err(); err != nil {
    return nil, fmt.Errorf("error reading config: %w", err)
}
```

**Warning signs:** Config file with I/O errors returns empty action list instead of error.

**Sources:** [Go bufio Scanner documentation](https://pkg.go.dev/bufio#Scanner)

### Pitfall 7: os/exec Does Not Invoke Shell by Default

**What goes wrong:** Shell features (pipelines `|`, redirections `>`, glob patterns `*.txt`, environment variable expansion `$HOME`) don't work unless explicitly using a shell.

**Why it happens:** os/exec runs programs directly like C's exec(), not through a shell interpreter.

**How to avoid:**
- For `run` actions, explicitly invoke shell: `exec.Command("sh", "-c", command)`
- This matches user expectation that `run npm install` works like typing it in terminal

**Warning signs:** Commands with pipes/redirects fail with "no such file or directory" errors.

**Sources:** [Go os/exec documentation](https://pkg.go.dev/os/exec), [Some Useful Patterns for Go's os/exec](https://www.dolthub.com/blog/2022-11-28-go-os-exec-patterns/)

## Code Examples

Verified patterns from official sources:

### Reading Config File Line by Line

```go
// Source: Go stdlib bufio documentation
func parseConfigFile(path string) ([]Action, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("failed to open config: %w", err)
    }
    defer file.Close()

    var actions []Action
    scanner := bufio.NewScanner(file)
    lineNum := 0

    for scanner.Scan() {
        lineNum++
        line := strings.TrimSpace(scanner.Text())

        // Skip blank lines and comments
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        // Split into action and path (max 2 parts for "run command with spaces")
        parts := strings.SplitN(line, " ", 2)
        if len(parts) < 2 {
            return nil, fmt.Errorf("line %d: invalid syntax (expected: action path)", lineNum)
        }

        actionType := parts[0]
        path := strings.TrimSpace(parts[1])

        actions = append(actions, Action{
            Type: actionType,
            Path: path,
            Line: lineNum,
        })
    }

    // Critical: check for I/O errors
    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("error reading config: %w", err)
    }

    return actions, nil
}
```

### Copying File/Directory with otiai10/copy

```go
// Source: github.com/otiai10/copy documentation
import cp "github.com/otiai10/copy"

func copyPath(src, dest string) error {
    // Basic copy with default options (preserves permissions)
    // Handles both files and directories recursively
    if err := cp.Copy(src, dest); err != nil {
        return fmt.Errorf("failed to copy %s: %w", src, err)
    }
    return nil
}

// With options for fine-grained control:
func copyPathWithOptions(src, dest string) error {
    opt := cp.Options{
        PreserveTimes: true,  // Keep timestamps
        OnSymlink: func(src string) cp.SymlinkAction {
            return cp.Shallow  // Copy symlinks as symlinks
        },
    }

    if err := cp.Copy(src, dest, opt); err != nil {
        return fmt.Errorf("failed to copy %s: %w", src, err)
    }
    return nil
}
```

### Creating Symlinks with Parent Directory Creation

```go
// Source: Go stdlib os documentation
func createSymlink(src, dest string) error {
    // Create parent directories if needed (like mkdir -p)
    parentDir := filepath.Dir(dest)
    if err := os.MkdirAll(parentDir, 0755); err != nil {
        return fmt.Errorf("failed to create parent directories: %w", err)
    }

    // Create symlink
    if err := os.Symlink(src, dest); err != nil {
        return fmt.Errorf("failed to create symlink: %w", err)
    }

    return nil
}
```

### Executing Command with Real-Time Output Streaming

```go
// Source: Go os/exec documentation
func executeCommand(workDir, command string) error {
    // Use shell to support pipes, redirects, etc.
    cmd := exec.Command("sh", "-c", command)
    cmd.Dir = workDir

    // Stream output to user in real-time
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Run(); err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            return fmt.Errorf("command exited with code %d", exitErr.ExitCode())
        }
        return fmt.Errorf("command failed: %w", err)
    }

    return nil
}
```

### Creating Parent Directories (mkdir -p equivalent)

```go
// Source: Go stdlib os documentation
func ensureParentDir(filePath string) error {
    parentDir := filepath.Dir(filePath)

    // MkdirAll is idempotent - returns nil if directory already exists
    if err := os.MkdirAll(parentDir, 0755); err != nil {
        return fmt.Errorf("failed to create parent directories: %w", err)
    }

    return nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Manual recursive copy with filepath.Walk | otiai10/copy library | Library released ~2018, widely adopted by 2020 | Eliminates 100+ lines of edge-case handling code |
| Cobra MarkFlagsMutuallyExclusive for complex groups | Custom validation in PreRunE | Cobra supports simple mutual exclusion but not "group vs group" | More flexible validation for config source model |
| bufio.Reader.ReadString('\n') | bufio.Scanner | Scanner API preferred since Go 1.1+ | Cleaner line-by-line semantics, better error handling |
| strings.Split for parsing | strings.SplitN with limit | SplitN available since Go 1.0 | Prevents over-splitting "run command with spaces" |

**Deprecated/outdated:**
- **ioutil package:** Deprecated in Go 1.16, functions moved to os and io packages. Use os.ReadFile, os.WriteFile, os.MkdirAll instead.
- **Manual symlink permission checks:** Windows 10+ Developer Mode makes symlinks available to non-admin users. Old workarounds (copying instead of symlinking on Windows) no longer needed.

## Open Questions

Things that couldn't be fully resolved:

1. **Should copy preserve timestamps by default?**
   - What we know: otiai10/copy provides PreserveTimes option, default is false
   - What's unclear: User expectation for worktree setup - do they expect timestamps to match source or reflect when worktree was created?
   - Recommendation: Default to false (new worktree = new timestamps), document option for Phase 5 if users request it

2. **How to handle symlinks inside copied directories?**
   - What we know: otiai10/copy supports 3 strategies (Deep, Shallow, Skip)
   - What's unclear: User expectation when copying a directory containing symlinks
   - Recommendation: Use default (Deep = copy symlink contents), matches behavior of `cp -r` without `-P` flag

3. **Should --run flags support multi-line commands?**
   - What we know: Config file supports multi-line via multiple `run` entries
   - What's unclear: Cobra StringSlice collects `--run "cmd1" --run "cmd2"` as separate entries. Should we support `--run "cmd1 && cmd2"`?
   - Recommendation: Both work naturally - StringSlice captures each flag as separate command, execute sequentially

## Sources

### Primary (HIGH confidence)

- [Go bufio package documentation](https://pkg.go.dev/bufio) - Scanner API and patterns
- [Go os package documentation](https://pkg.go.dev/os) - File operations, MkdirAll, symlinks
- [Go os/exec package documentation](https://pkg.go.dev/os/exec) - Command execution
- [Go strings package documentation](https://pkg.go.dev/strings) - String manipulation
- [github.com/otiai10/copy v1.14.1](https://pkg.go.dev/github.com/otiai10/copy) - Recursive copy library
- [Cobra documentation: Working with Flags](https://cobra.dev/docs/how-to-guides/working-with-flags/) - Flag validation patterns
- [Go database transactions documentation](https://go.dev/doc/database/execute-transactions) - Rollback pattern (official Go pattern)

### Secondary (MEDIUM confidence)

- [Advanced command execution in Go with os/exec](https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html) - Streaming output patterns
- [Reading a File Line by Line in Go](https://leapcell.io/blog/reading-a-file-line-by-line-in-go) - bufio.Scanner best practices
- [Creating Directories with os.Mkdir() and os.MkdirAll()](https://reintech.io/term/creating-directories-with-os-mkdir-and-os-mkdirall-in-go) - Parent directory creation
- [Some Useful Patterns for Go's os/exec](https://www.dolthub.com/blog/2022-11-28-go-os-exec-patterns/) - Shell invocation patterns
- [Database Transactions in Go with Layered Architecture](https://threedots.tech/post/database-transactions-in-go/) - Deferred rollback pattern
- [Cross-platform symlinks](https://geedew.com/cross-platform-symlink/) - Windows symlink behavior
- [Understanding os.Stat() vs os.Lstat()](https://dev.to/moseeh_52/understanding-osstat-vs-oslstat-in-go-file-and-symlink-handling-3p5d) - Symlink validation

### Tertiary (LOW confidence)

- Web search results for Go config parsing best practices (generic advice, no specific 2026 updates)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All stdlib + one widely-used library (1,264+ dependents)
- Architecture: HIGH - Patterns verified from official Go documentation and established codebases
- Pitfalls: HIGH - Sourced from Go issue tracker and official documentation warnings

**Research date:** 2026-02-07
**Valid until:** ~60 days (stable stdlib, mature library ecosystem, slow-moving domain)
