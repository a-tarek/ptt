# Architecture Patterns

**Domain:** Go CLI tool with npm binary distribution and shell integration
**Researched:** 2026-02-07

## Recommended Architecture

The `wt` rewrite follows a **three-layer architecture**:

```
┌─────────────────────────────────────────────────────────────┐
│ Shell Layer (bash/zsh/fish)                                 │
│ - Thin wrapper functions sourced into user's shell          │
│ - Handles `cd` operations (can't be done by subprocess)     │
│ - Invokes Go binary, captures output, acts on it            │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Go Binary Layer                                             │
│ - All business logic (git operations, validation, etc.)     │
│ - Outputs structured response (path to cd, or error)        │
│ - No side effects on parent process                         │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ npm Distribution Layer                                      │
│ - Main package (@scope/wt) with postinstall script          │
│ - Platform-specific optional dependencies                   │
│ - Interactive installer for shell setup                     │
└─────────────────────────────────────────────────────────────┘
```

### Component Boundaries

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| **Shell wrapper** | Directory changes, environment interaction | Go binary via exec |
| **Go binary** | Git logic, validation, config parsing | Git, filesystem |
| **npm package** | Binary distribution, installation | OS package managers, filesystem |
| **Installer script** | Shell detection, dotfile modification | Shell rc files, filesystem |

### Data Flow

```
User types: wt goto staging

1. Shell wrapper function `wt()` receives "goto staging"
2. Shell wrapper checks if command needs cd (goto/new/home/eject)
3. For cd commands:
   a. Invoke: wt-bin goto staging --output-path
   b. Binary validates, resolves worktree, outputs: /path/to/worktree
   c. Shell wrapper: cd /path/to/worktree
4. For non-cd commands:
   a. Invoke: wt-bin list (or merge/rebase/delete/init)
   b. Binary executes, outputs to stdout/stderr
   c. Shell wrapper passes through
```

**Key insight:** The binary never attempts to `cd` — it only outputs paths. The shell wrapper is the only component that changes directories.

## Go Project Structure

### Standard Layout (cmd/internal/pkg pattern)

```
wt/
├── cmd/
│   └── wt/                    # Main package
│       └── main.go            # Entry point, cobra root command setup
├── internal/                  # Private application code
│   ├── commands/              # Cobra command implementations
│   │   ├── new.go
│   │   ├── goto.go
│   │   ├── home.go
│   │   ├── init.go
│   │   ├── eject.go
│   │   ├── list.go
│   │   ├── merge.go
│   │   ├── rebase.go
│   │   └── delete.go
│   ├── git/                   # Git operations abstraction
│   │   ├── worktree.go        # Worktree operations
│   │   ├── repository.go      # Repository queries
│   │   └── branch.go          # Branch operations
│   ├── config/                # .wtconfig parsing and application
│   │   ├── parser.go          # Parse .wtconfig file
│   │   └── setup.go           # Apply copy/symlink actions
│   ├── resolver/              # Worktree name/path/branch resolution
│   │   └── resolver.go
│   └── output/                # Structured output formatting
│       └── formatter.go       # --output-path vs human-readable
├── pkg/                       # Public libraries (if any)
│   └── wtconfig/              # Config file types (exportable)
├── scripts/                   # Build and distribution scripts
│   ├── build.sh               # Cross-platform build script
│   └── install.js             # npm postinstall interactive installer
├── shell/                     # Shell wrapper functions
│   ├── wt.bash
│   ├── wt.zsh
│   └── wt.fish
├── completions/               # Shell completions
│   ├── wt.bash
│   ├── wt.zsh
│   └── wt.fish
├── npm/                       # npm package structure
│   ├── package.json           # Main package
│   └── platforms/             # Platform-specific packages
│       ├── darwin-arm64/
│       │   └── package.json
│       ├── darwin-x64/
│       │   └── package.json
│       ├── linux-x64/
│       │   └── package.json
│       └── win32-x64/
│           └── package.json
├── go.mod
├── go.sum
└── README.md
```

### Directory Rationale

**`cmd/wt/`**: Single entry point. Initializes cobra root command and executes.

**`internal/commands/`**: One file per cobra command. Each exports a function returning `*cobra.Command`. Keeps command logic isolated and testable.

**`internal/git/`**: Abstracts all `git` CLI invocations. Makes testing easier (mock git operations). Single source of truth for git command construction.

**`internal/config/`**: Handles `.wtconfig` parsing and application. Separated because it's complex (overrides, validation) and reused by multiple commands (new, eject).

**`internal/resolver/`**: Name-to-path and name-to-branch resolution. Used by goto, merge, rebase, delete. Centralized logic prevents duplication.

**`internal/output/`**: Handles two output modes:
- `--output-path`: Machine-readable (just the path, for shell wrapper)
- Default: Human-readable (colorized, multi-line)

**`pkg/`**: Exportable types (if needed). Start with none; add only if external tools need to parse `.wtconfig`.

**`shell/`**: Thin wrapper functions. These are **not built into the binary** — they're separate files installed by the npm package.

**`npm/`**: npm package structure with platform-specific optional dependencies pattern.

## Cobra Command Structure

### Root Command (cmd/wt/main.go)

```go
package main

import (
    "os"
    "github.com/spf13/cobra"
    "github.com/user/wt/internal/commands"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "wt",
        Short: "Git worktree manager",
        Long:  "A fast, ergonomic Git worktree manager",
    }

    // Add subcommands
    rootCmd.AddCommand(
        commands.NewCmd(),
        commands.GotoCmd(),
        commands.HomeCmd(),
        commands.InitCmd(),
        commands.EjectCmd(),
        commands.ListCmd(),
        commands.MergeCmd(),
        commands.RebaseCmd(),
        commands.DeleteCmd(),
    )

    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Command Implementation Pattern (internal/commands/goto.go)

```go
package commands

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/user/wt/internal/git"
    "github.com/user/wt/internal/output"
    "github.com/user/wt/internal/resolver"
)

func GotoCmd() *cobra.Command {
    var outputPath bool

    cmd := &cobra.Command{
        Use:   "goto <worktree>",
        Short: "Change directory to a worktree",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            name := args[0]

            // Resolve worktree name to path
            path, err := resolver.ResolvePath(name)
            if err != nil {
                return fmt.Errorf("worktree '%s' not found", name)
            }

            // Output mode: just path (for shell wrapper) or human-readable
            if outputPath {
                fmt.Println(path)
            } else {
                output.Success("Changed to: %s", path)
            }

            return nil
        },
    }

    cmd.Flags().BoolVar(&outputPath, "output-path", false, "Output only the path (for shell integration)")

    return cmd
}
```

### Commands Needing `--output-path` Flag

| Command | Needs cd | Output Format |
|---------|----------|---------------|
| `new` | YES | `--output-path` → path only |
| `goto` | YES | `--output-path` → path only |
| `home` | YES | `--output-path` → path only |
| `eject` | YES | `--output-path` → path only |
| `list` | NO | Human-readable table |
| `merge` | NO | Git output passthrough |
| `rebase` | NO | Git output passthrough |
| `delete` | NO | Confirmation message |
| `init` | NO | Confirmation message |

## Shell Wrapper Architecture

### The Problem

A subprocess cannot change its parent's directory. When the Go binary executes `cd`, it only affects its own process, not the user's shell.

### The Solution

Thin shell functions sourced into the user's shell. These functions:
1. Invoke the Go binary with `--output-path` for cd commands
2. Capture stdout (the path)
3. Execute `cd` in the current shell context

### Wrapper Implementation (shell/wt.zsh)

```zsh
#!/usr/bin/env zsh
# wt shell wrapper for zsh

function wt() {
  local cmd="$1"
  shift 2>/dev/null

  case "$cmd" in
    new|goto|home|eject)
      # Commands that need cd
      local target_path
      target_path=$(wt-bin "$cmd" --output-path "$@" 2>&1)
      local exit_code=$?

      if [[ $exit_code -eq 0 && -n "$target_path" ]]; then
        cd "$target_path"
      else
        # Error occurred — output was error message, not path
        echo "$target_path" >&2
        return $exit_code
      fi
      ;;
    *)
      # Commands that don't need cd — pass through
      wt-bin "$cmd" "$@"
      ;;
  esac
}
```

### Key Design Decisions

1. **Command categorization**: Hardcoded list of cd commands in wrapper. Alternative (flag-based detection) is more fragile.

2. **Error handling**: If binary exits non-zero, output is treated as error message, not path.

3. **Stdout capture**: Only stdout is captured for path. Stderr passes through for progress messages during long operations.

4. **Binary name**: `wt-bin` to distinguish from shell function `wt`. Prevents infinite recursion.

## npm Package Structure

### Pattern: Platform-Specific Optional Dependencies

This is the standard pattern used by esbuild, swc, prisma, and other tools distributing native binaries via npm.

### Main Package (@scope/wt)

```json
{
  "name": "@scope/wt",
  "version": "1.0.0",
  "description": "Git worktree manager",
  "bin": {
    "wt-bin": "./bin/wt"
  },
  "optionalDependencies": {
    "@scope/wt-darwin-arm64": "1.0.0",
    "@scope/wt-darwin-x64": "1.0.0",
    "@scope/wt-linux-x64": "1.0.0",
    "@scope/wt-linux-arm64": "1.0.0",
    "@scope/wt-win32-x64": "1.0.0"
  },
  "scripts": {
    "postinstall": "node scripts/install.js"
  },
  "files": [
    "bin/",
    "shell/",
    "completions/",
    "scripts/"
  ]
}
```

### Platform-Specific Package (@scope/wt-darwin-arm64)

```json
{
  "name": "@scope/wt-darwin-arm64",
  "version": "1.0.0",
  "os": ["darwin"],
  "cpu": ["arm64"],
  "files": [
    "wt"
  ]
}
```

Each platform package contains a single file: the compiled binary for that platform.

### How npm Resolves Binaries

1. User runs `npm install @scope/wt`
2. npm installs main package
3. npm attempts to install all `optionalDependencies`
4. npm skips packages where `os`/`cpu` don't match current platform
5. **Only the matching platform package installs**
6. Postinstall script runs

### Postinstall Script (scripts/install.js)

```javascript
#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const os = require('os');

// Determine which platform package was installed
const platform = `${process.platform}-${process.arch}`;
const binarySource = path.join(
  __dirname,
  '..',
  'node_modules',
  `@scope/wt-${platform}`,
  'wt'
);

// Copy binary to bin/ directory
const binaryDest = path.join(__dirname, '..', 'bin', 'wt');

if (fs.existsSync(binarySource)) {
  fs.mkdirSync(path.dirname(binaryDest), { recursive: true });
  fs.copyFileSync(binarySource, binaryDest);
  fs.chmodSync(binaryDest, 0o755);

  // Run interactive installer
  require('./interactive-installer.js');
} else {
  console.error(`Error: No binary found for platform ${platform}`);
  console.error('Supported platforms: darwin-arm64, darwin-x64, linux-x64, linux-arm64, win32-x64');
  process.exit(1);
}
```

### Why Not Just Include All Binaries?

Including all binaries in the main package would work but:
- Increases package size 5x (download bloat)
- npm's optional dependencies pattern is the ecosystem standard
- Tools expect this pattern (security scanners, CI caches)

## Interactive Installer Architecture

The installer runs during `npm install` (postinstall hook) and sets up shell integration.

### Goals

1. Detect user's shell (bash/zsh/fish)
2. Determine correct rc file (~/.bashrc, ~/.zshrc, ~/.config/fish/config.fish)
3. Check if already installed (idempotent)
4. Prompt user for confirmation
5. Append source line to rc file
6. Install shell completions

### Installer Flow

```
npm install @scope/wt
  ↓
postinstall hook runs
  ↓
Copy binary from platform package
  ↓
Run interactive installer
  ↓
┌─────────────────────────────────────┐
│ Detect shell (SHELL env var)       │
└──────────────┬──────────────────────┘
               ↓
┌─────────────────────────────────────┐
│ Determine rc file location          │
│ - bash: ~/.bashrc or ~/.bash_profile│
│ - zsh: ~/.zshrc                     │
│ - fish: ~/.config/fish/config.fish  │
└──────────────┬──────────────────────┘
               ↓
┌─────────────────────────────────────┐
│ Check if already installed          │
│ (search for wt marker in rc file)   │
└──────────────┬──────────────────────┘
               ↓
          Already installed?
               ├─ YES → Exit (idempotent)
               └─ NO → Continue
                       ↓
               ┌─────────────────────────┐
               │ Prompt user:            │
               │ "Add wt to ~/.zshrc?"   │
               │ [Y/n]                   │
               └──────┬──────────────────┘
                      ↓
                User confirms?
                      ├─ NO → Exit
                      └─ YES → Continue
                              ↓
                      ┌─────────────────────┐
                      │ Append to rc file:  │
                      │ source wt.zsh       │
                      │ source wt-comp.zsh  │
                      └──────┬──────────────┘
                             ↓
                      ┌─────────────────────┐
                      │ Success message:    │
                      │ "Restart shell or:  │
                      │  source ~/.zshrc"   │
                      └─────────────────────┘
```

### Implementation (scripts/interactive-installer.js)

```javascript
const readline = require('readline');
const fs = require('fs');
const path = require('path');
const os = require('os');

function detectShell() {
  const shell = process.env.SHELL || '';
  if (shell.includes('zsh')) return 'zsh';
  if (shell.includes('bash')) return 'bash';
  if (shell.includes('fish')) return 'fish';
  return null;
}

function getRcFile(shell) {
  const home = os.homedir();
  switch (shell) {
    case 'zsh':
      return path.join(home, '.zshrc');
    case 'bash':
      // Prefer .bashrc on Linux, .bash_profile on macOS
      const bashrc = path.join(home, '.bashrc');
      const bashProfile = path.join(home, '.bash_profile');
      return fs.existsSync(bashrc) ? bashrc : bashProfile;
    case 'fish':
      return path.join(home, '.config', 'fish', 'config.fish');
    default:
      return null;
  }
}

function isInstalled(rcFile) {
  if (!fs.existsSync(rcFile)) return false;
  const content = fs.readFileSync(rcFile, 'utf8');
  return content.includes('# wt - Git Worktree Manager');
}

async function promptUser(question) {
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
  });

  return new Promise((resolve) => {
    rl.question(question, (answer) => {
      rl.close();
      resolve(answer.trim().toLowerCase());
    });
  });
}

async function install() {
  console.log('\n🔧 Setting up wt shell integration...\n');

  const shell = detectShell();
  if (!shell) {
    console.log('⚠️  Could not detect shell. Manual setup required.');
    console.log('Add this to your shell rc file:');
    console.log(`  source ${path.join(__dirname, '..', 'shell', 'wt.<shell>')}`);
    return;
  }

  const rcFile = getRcFile(shell);
  if (!rcFile) {
    console.log('⚠️  Could not determine rc file location.');
    return;
  }

  if (isInstalled(rcFile)) {
    console.log('✅ wt is already installed in', rcFile);
    return;
  }

  console.log(`Detected shell: ${shell}`);
  console.log(`RC file: ${rcFile}`);
  console.log('');

  const answer = await promptUser(`Add wt to ${rcFile}? [Y/n] `);
  if (answer && answer !== 'y' && answer !== 'yes') {
    console.log('Skipped. You can install manually later.');
    return;
  }

  // Append to rc file
  const wrapperPath = path.join(__dirname, '..', 'shell', `wt.${shell}`);
  const completionPath = path.join(__dirname, '..', 'completions', `wt.${shell}`);

  const snippet = `
# wt - Git Worktree Manager
source "${wrapperPath}"
source "${completionPath}"
`;

  fs.appendFileSync(rcFile, snippet);

  console.log('✅ Successfully installed!');
  console.log('');
  console.log('Restart your shell or run:');
  console.log(`  source ${rcFile}`);
}

install().catch(console.error);
```

### Idempotence Strategy

Installer checks for marker comment `# wt - Git Worktree Manager` in rc file. If found, skips installation. This allows:
- Re-running `npm install` without duplication
- Uninstall/reinstall workflows
- Version upgrades

### Manual Setup Fallback

If shell detection fails or user declines, installer shows manual instructions. Critical for:
- Unsupported shells
- Custom shell configurations
- CI/CD environments

## Patterns to Follow

### Pattern 1: Dependency Inversion for Git Operations

**What:** Abstract git CLI calls behind an interface

**When:** Testing, mocking git behavior

**Example:**
```go
// internal/git/interface.go
type Repository interface {
    WorktreeList() ([]Worktree, error)
    WorktreeAdd(path, branch string) error
    WorktreeRemove(path string) error
}

// internal/git/gitcli.go
type GitCLI struct{}

func (g *GitCLI) WorktreeList() ([]Worktree, error) {
    // exec git worktree list --porcelain
}

// In tests: mock Repository
type MockRepository struct {
    worktrees []Worktree
}
```

### Pattern 2: Structured Output Mode

**What:** Two output modes — machine-readable and human-readable

**When:** Shell integration commands (new, goto, home, eject)

**Example:**
```go
// internal/output/formatter.go
type Formatter struct {
    PathOnly bool
}

func (f *Formatter) Success(path string, message string) {
    if f.PathOnly {
        fmt.Println(path)
    } else {
        fmt.Printf("✅ %s\n", message)
        fmt.Printf("📁 %s\n", path)
    }
}
```

### Pattern 3: Config Override Chain

**What:** Merge .wtconfig entries with CLI flag overrides

**When:** `wt new --copy .env` should override `.wtconfig` entry

**Example:**
```go
// internal/config/setup.go
func ApplySetup(src, dest string, overrides map[string]Action) error {
    config := parseConfig(path.Join(src, ".wtconfig"))

    // Apply overrides
    for path, action := range overrides {
        config[path] = action
    }

    for path, action := range config {
        switch action {
        case Copy:
            copyFile(path.Join(src, path), path.Join(dest, path))
        case Symlink:
            symlinkFile(path.Join(src, path), path.Join(dest, path))
        }
    }
}
```

### Pattern 4: Cobra Command Factory Functions

**What:** Each command is a function returning `*cobra.Command`

**When:** Setting up root command

**Why:** Allows command-specific flag variables without global state

**Example:**
```go
// internal/commands/new.go
func NewCmd() *cobra.Command {
    var (
        outputPath bool
        copyPaths  []string
        symlinkPaths []string
    )

    cmd := &cobra.Command{
        Use: "new",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Use local variables
        },
    }

    cmd.Flags().BoolVar(&outputPath, "output-path", false, "...")
    cmd.Flags().StringArrayVar(&copyPaths, "copy", nil, "...")
    cmd.Flags().StringArrayVar(&symlinkPaths, "symlink", nil, "...")

    return cmd
}
```

## Anti-Patterns to Avoid

### Anti-Pattern 1: Binary Attempts to cd

**What:** Go binary tries to change user's directory

**Why bad:** Subprocesses cannot affect parent shell

**Instead:** Binary outputs path, shell wrapper does cd

### Anti-Pattern 2: Monolithic main.go

**What:** All command logic in cmd/wt/main.go

**Why bad:**
- Untestable (main package can't be imported)
- Difficult to navigate
- Violates single responsibility

**Instead:** Move logic to internal/commands/, keep main.go minimal

### Anti-Pattern 3: Global State for Flags

**What:** Package-level variables for cobra flags

```go
// BAD
var outputPath bool

func init() {
    rootCmd.PersistentFlags().BoolVar(&outputPath, "output-path", false, "...")
}
```

**Why bad:**
- Makes testing difficult
- Causes race conditions in parallel tests
- Unclear ownership

**Instead:** Use command factory pattern with local variables

### Anti-Pattern 4: Bundling All Binaries in One npm Package

**What:** Include darwin-arm64, darwin-x64, linux-x64, win32-x64 binaries in single package

**Why bad:**
- 50-100MB download for 10-20MB needed binary
- Wastes bandwidth, disk space
- Violates npm ecosystem conventions

**Instead:** Use platform-specific optional dependencies

### Anti-Pattern 5: Silent Postinstall

**What:** Modify shell rc files without user confirmation

**Why bad:**
- Violates user expectations
- npm install should be non-interactive (CI/CD)
- Security concern (arbitrary shell code injection)

**Instead:** Prompt for confirmation, provide manual fallback

### Anti-Pattern 6: Assuming Shell Type

**What:** Only support zsh, ignore bash/fish

**Why bad:**
- Alienates large user base
- Prevents adoption in teams with diverse setups
- Creates compatibility issues

**Instead:** Detect shell, provide wrappers for bash/zsh/fish

## Scalability Considerations

| Concern | At 10 users | At 1K users | At 100K users |
|---------|------------|-------------|---------------|
| Binary size | <10MB acceptable | <10MB still fine | <5MB ideal (optimize) |
| Platform support | macOS/Linux sufficient | Add Windows | Add ARM variants |
| Shell support | zsh only | bash + zsh | bash + zsh + fish |
| Installation UX | Manual setup OK | Need npm package | Polish installer, error handling |
| Documentation | README sufficient | Need examples, troubleshooting | Video tutorials, community support |
| Error messages | Generic errors OK | Context-specific errors | Actionable suggestions |

## Build Order Dependencies

Suggested implementation order based on dependencies:

### Phase 1: Core Go Binary (no cd commands)

1. **Project setup**: `go mod init`, directory structure
2. **Git abstraction**: `internal/git/` (worktree list, branch operations)
3. **Simple commands**: `list`, `init`, `delete`
4. **Cobra structure**: Root command, subcommand registration
5. **Testing**: Unit tests for git abstraction

**Milestone**: `wt-bin list` works, outputs worktree table

### Phase 2: Config System

1. **Parser**: `internal/config/parser.go` (.wtconfig syntax)
2. **Setup logic**: `internal/config/setup.go` (copy/symlink application)
3. **Override support**: CLI flags override config entries
4. **Testing**: Config parsing edge cases

**Milestone**: Can parse .wtconfig, apply actions (tested separately)

### Phase 3: Complex Commands

1. **Resolver**: `internal/resolver/` (name → path, name → branch)
2. **Commands with cd**: `new`, `goto`, `home`, `eject`
3. **Output formatting**: `--output-path` flag support
4. **Testing**: Integration tests with temp git repos

**Milestone**: `wt-bin new test --output-path` outputs path

### Phase 4: Shell Integration

1. **Shell wrappers**: `shell/wt.{bash,zsh,fish}`
2. **Completions**: `completions/wt.{bash,zsh,fish}`
3. **Manual testing**: Source wrapper, test cd commands
4. **Edge cases**: Error handling, path with spaces

**Milestone**: Can `source shell/wt.zsh` and `wt goto` changes directory

### Phase 5: npm Distribution

1. **Build script**: `scripts/build.sh` (cross-compile for all platforms)
2. **npm packages**: Main package + platform-specific packages
3. **Postinstall**: Binary copy script
4. **Testing**: `npm pack`, test on different platforms

**Milestone**: Can `npm install`, binary is available as `wt-bin`

### Phase 6: Interactive Installer

1. **Shell detection**: Detect bash/zsh/fish from $SHELL
2. **RC file logic**: Determine correct rc file path
3. **Idempotence**: Check if already installed
4. **Interactive prompts**: User confirmation
5. **Append logic**: Add source lines to rc file
6. **Testing**: Test on different shells, edge cases

**Milestone**: `npm install` prompts and sets up shell integration

### Phase 7: Polish & Documentation

1. **Error messages**: Improve UX, add suggestions
2. **README**: Installation, usage, examples
3. **Troubleshooting**: Common issues, manual setup
4. **Examples**: Workflow guides
5. **CI/CD**: GitHub Actions for releases

**Milestone**: Production-ready, documented, tested

### Critical Path

```
Core Go Binary → Config System → Complex Commands → Shell Integration
                                                              ↓
                                                   npm Distribution
                                                              ↓
                                                   Interactive Installer
```

**Cannot parallelize:**
- Shell wrappers need `--output-path` support in binary
- npm distribution needs compiled binaries
- Installer needs shell wrappers and completions

**Can parallelize:**
- Completions and shell wrappers (independent)
- Documentation and installer (independent)

## Detailed Component Diagrams

### npm Package Dependency Graph

```
@scope/wt (main package)
├── optionalDependencies
│   ├── @scope/wt-darwin-arm64 ← Only installs on macOS ARM64
│   ├── @scope/wt-darwin-x64   ← Only installs on macOS x64
│   ├── @scope/wt-linux-x64    ← Only installs on Linux x64
│   ├── @scope/wt-linux-arm64  ← Only installs on Linux ARM64
│   └── @scope/wt-win32-x64    ← Only installs on Windows x64
└── files
    ├── bin/wt                 ← Symlink to installed binary
    ├── shell/
    │   ├── wt.bash
    │   ├── wt.zsh
    │   └── wt.fish
    ├── completions/
    │   ├── wt.bash
    │   ├── wt.zsh
    │   └── wt.fish
    └── scripts/
        ├── install.js         ← Postinstall: copy binary from platform package
        └── interactive-installer.js
```

### Go Module Dependency Graph

```
cmd/wt/main.go
└── internal/commands/*
    ├── internal/git
    ├── internal/config
    ├── internal/resolver
    │   └── internal/git
    └── internal/output
```

No circular dependencies. All dependencies flow downward.

## Testing Strategy

### Unit Tests

- `internal/git`: Mock git CLI, test parsing
- `internal/config`: Test .wtconfig parser with various inputs
- `internal/resolver`: Mock worktree list, test name resolution

### Integration Tests

- Create temp git repos with worktrees
- Run commands, verify filesystem changes
- Test error conditions (missing branch, invalid path)

### Shell Integration Tests

- Source shell wrapper in test shell
- Execute commands, verify `$PWD` changes
- Test error passthrough

### npm Package Tests

- `npm pack` → install in temp directory
- Verify binary is executable
- Verify postinstall runs
- Test on different platforms (CI matrix)

## Sources

This architecture document is based on established patterns from:

- **Go project layout**: golang-standards/project-layout (community standard)
- **Cobra CLI**: spf13/cobra (used by kubectl, GitHub CLI, Hugo)
- **npm binary distribution**: esbuild, swc, prisma (platform-specific optional dependencies pattern)
- **Shell integration**: nvm, rvm, z.sh (source-based wrappers for cd functionality)

**Confidence level: HIGH** - These are well-established, battle-tested patterns used by major tools in the ecosystem.
