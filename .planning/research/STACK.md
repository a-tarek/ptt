# Technology Stack

**Project:** wt (Git Worktree Manager - Go Rewrite)
**Researched:** 2026-02-07
**Confidence:** MEDIUM (based on training data through Jan 2025; versions should be verified)

## Recommended Stack

### Core Framework
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.21+ | Core language | Latest stable with generics, improved error handling. Use 1.22+ if available for enhanced routing performance |
| cobra | v1.8+ | CLI framework & command structure | De facto standard for Go CLIs. Built-in completion generation for bash/zsh/fish/powershell. Used by kubectl, gh, hugo |
| pflag | v1.0.5+ | POSIX/GNU-style flags | Cobra dependency. Superior to stdlib flag package for CLI apps |

### Build & Release
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| goreleaser | v1.23+ | Multi-platform binary builds & npm publishing | Industry standard. Native npm publisher support. Handles cross-compilation, checksums, archives |
| GoReleaser Pro | (optional) | Advanced features | Only if need custom npm scoping, advanced hooks. OSS version sufficient for this project |
| Makefile | - | Local dev builds | Standard for Go projects. Quick iteration without full release flow |

### Testing & Quality
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| testing (stdlib) | - | Unit tests | Sufficient for CLI testing. No need for external framework |
| testify/assert | v1.8+ | Test assertions | Optional but improves readability over plain if-checks |
| golangci-lint | v1.55+ | Linting | Meta-linter. Runs multiple linters (go vet, staticcheck, etc.) |

### npm Distribution
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| npm (package.json) | - | Package metadata | Standard npm package structure |
| node (package.json) | >=16 | Runtime requirement | For npx and postinstall scripts |
| Platform packages | - | OS/arch-specific binaries | Separate npm packages per platform (see Architecture section) |

## Architecture: npm Binary Distribution Pattern

### Pattern: esbuild/turbo/biome Style

**Structure:**
```
@scope/wt/                   # Main wrapper package (small, OS-agnostic)
  package.json              # Depends on optionalDependencies for platforms
  bin/wt.js                 # Node wrapper that finds/executes binary
  install.js                # Interactive installer (npx @scope/wt install)

@scope/wt-darwin-x64/       # Platform-specific package
  package.json
  bin/wt                    # Actual Go binary

@scope/wt-darwin-arm64/
@scope/wt-linux-x64/
@scope/wt-linux-arm64/
@scope/wt-windows-x64/
... (one per GOOS/GOARCH combo)
```

**Why this pattern:**
- npm only downloads binary for user's platform (via optionalDependencies)
- Fast installs (only 1 binary + small wrapper, not all platforms)
- Works with npx for one-off execution
- Proven by esbuild (100M+ weekly downloads), biome, turbo

**Main package.json:**
```json
{
  "name": "@scope/wt",
  "version": "0.1.0",
  "bin": {
    "wt": "./bin/wt.js"
  },
  "optionalDependencies": {
    "@scope/wt-darwin-x64": "0.1.0",
    "@scope/wt-darwin-arm64": "0.1.0",
    "@scope/wt-linux-x64": "0.1.0",
    "@scope/wt-linux-arm64": "0.1.0",
    "@scope/wt-windows-x64": "0.1.0"
  }
}
```

**bin/wt.js (Node wrapper):**
```javascript
#!/usr/bin/env node
const { execFileSync } = require('child_process');
const { join } = require('path');
const { platform, arch } = process;

// Map Node platform/arch to Go GOOS/GOARCH
const PLATFORMS = {
  'darwin-x64': '@scope/wt-darwin-x64',
  'darwin-arm64': '@scope/wt-darwin-arm64',
  'linux-x64': '@scope/wt-linux-x64',
  'linux-arm64': '@scope/wt-linux-arm64',
  'win32-x64': '@scope/wt-windows-x64',
};

const platformKey = `${platform}-${arch}`;
const packageName = PLATFORMS[platformKey];

if (!packageName) {
  console.error(`Unsupported platform: ${platformKey}`);
  process.exit(1);
}

try {
  const binaryPath = require.resolve(`${packageName}/bin/wt`);
  execFileSync(binaryPath, process.argv.slice(2), { stdio: 'inherit' });
} catch (err) {
  console.error(`Failed to execute wt binary: ${err.message}`);
  process.exit(1);
}
```

**Why NOT postinstall scripts:**
- Blocked by many corporate npm configs
- Unreliable in CI/CD environments
- optionalDependencies + wrapper is more robust

## Shell Wrapper Generation

### Approach: Cobra Completions + Custom Wrapper Template

**Cobra provides:**
- `cobra.Command.GenBashCompletion()` - bash completion script
- `cobra.Command.GenZshCompletion()` - zsh completion script
- `cobra.Command.GenFishCompletion()` - fish completion script
- `cobra.Command.GenPowerShellCompletion()` - powershell completion

**Custom wrapper needed for directory-changing commands:**

Since `goto`, `home`, `new`, `eject` need to change parent shell's directory, Go binary outputs:
```
CHDIR:/path/to/worktree
```

Shell wrapper sources this and executes cd:

**bash/zsh wrapper (~/.local/bin/wt):**
```bash
#!/bin/bash
wt() {
  local output=$(command wt "$@")
  if [[ "$output" =~ ^CHDIR:(.*)$ ]]; then
    cd "${BASH_REMATCH[1]}" || return 1
  else
    echo "$output"
  fi
}
```

**install.js interactive installer:**
```javascript
// Prompts:
// - Which shell? (detect from $SHELL, allow override)
// - Where to install wrapper? (default ~/.local/bin or ~/bin)
// - Add to PATH in shell rc? (optional)
// - Install completions? (default yes)

// Actions:
// 1. Run `wt completion bash|zsh|fish` to generate completion script
// 2. Generate shell wrapper with cd support
// 3. Install to chosen location
// 4. Optionally append to .bashrc/.zshrc/.config/fish/config.fish
```

## Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| go-git/go-git | v5.11+ | Pure Go git implementation | If need programmatic git operations. Alternative: shell out to git binary (simpler, more reliable) |
| fatih/color | v1.16+ | Colored terminal output | For user-friendly error/success messages |
| AlecAivazis/survey | v2.3+ | Interactive prompts | If interactive mode beyond install (e.g., `wt new --interactive`) |
| spf13/viper | v1.18+ | Configuration management | If need config file support (likely not needed for wt) |

**Recommendation for wt:**
- DO use: cobra, pflag, goreleaser
- DO use: color (for better UX)
- DO NOT use: go-git (shell out to git binary instead - simpler, more reliable, respects user's git config)
- DO NOT use: survey (installer handles prompts in Node.js, not Go binary)
- DO NOT use: viper (no config file needed; state is in git worktree structure)

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| CLI Framework | cobra | urfave/cli | cobra has better completion generation, more active maintenance, industry standard |
| Build/Release | goreleaser | Manual Makefile + npm publish | goreleaser handles checksums, multi-platform, npm publish in one config |
| npm Pattern | optionalDependencies | Single fat package with all binaries | Would be 50MB+ download vs 5MB; esbuild pattern is proven |
| npm Pattern | optionalDependencies | postinstall script downloads binary | Blocked by corporate firewalls; unreliable; security concerns |
| Shell Integration | Wrapper function | Alias with eval | Wrapper is more flexible, allows cd support |
| Git Integration | Shell out to git binary | go-git library | Simpler, respects user's git config, smaller binary size |

## Installation & Setup

### Local Development

```bash
# Install Go 1.21+
brew install go  # macOS
# or download from golang.org

# Initialize Go module
go mod init github.com/yourusername/wt
go get github.com/spf13/cobra@latest
go get github.com/spf13/pflag@latest
go get github.com/fatih/color@latest

# Dev dependencies
brew install goreleaser golangci-lint

# Build locally
make build
# or
go build -o wt cmd/wt/main.go
```

### npm Publishing Setup

```bash
# Initialize npm packages
mkdir -p npm/wt npm/wt-darwin-x64 npm/wt-darwin-arm64 \
         npm/wt-linux-x64 npm/wt-linux-arm64 npm/wt-windows-x64

# Each platform package needs:
# - package.json with name, version
# - Binary in bin/ directory (goreleaser handles this)

# Main wrapper package needs:
# - package.json with optionalDependencies
# - bin/wt.js wrapper script
# - install.js interactive installer
```

### goreleaser Configuration

**.goreleaser.yml:**
```yaml
builds:
  - id: wt
    binary: wt
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - id: wt
    format: binary

publishers:
  - name: npm
    ids:
      - wt
    dir: "{{ dir .ArtifactPath }}"
    cmd: |
      # Custom script to:
      # 1. Copy binary to appropriate npm/wt-{os}-{arch}/bin/
      # 2. Update version in package.json
      # 3. Run npm publish
      ./scripts/publish-npm.sh {{ .Os }} {{ .Arch }} {{ .Version }}
```

## Project Structure

```
wt/
├── cmd/
│   └── wt/
│       └── main.go              # Entry point
├── internal/
│   ├── commands/
│   │   ├── root.go             # Root cobra command
│   │   ├── new.go              # wt new
│   │   ├── goto.go             # wt goto
│   │   ├── home.go             # wt home
│   │   ├── init.go             # wt init
│   │   ├── eject.go            # wt eject
│   │   ├── list.go             # wt list
│   │   ├── merge.go            # wt merge
│   │   ├── rebase.go           # wt rebase
│   │   ├── delete.go           # wt delete
│   │   └── completion.go       # wt completion (shell completion generation)
│   ├── worktree/
│   │   ├── worktree.go         # Worktree operations
│   │   └── config.go           # Configuration (wt.cfg)
│   └── git/
│       └── git.go              # Git command wrappers
├── npm/
│   ├── wt/                     # Main wrapper package
│   │   ├── package.json
│   │   ├── bin/
│   │   │   └── wt.js          # Node wrapper
│   │   └── install.js         # Interactive installer
│   ├── wt-darwin-x64/
│   │   ├── package.json
│   │   └── bin/               # goreleaser puts binary here
│   ├── wt-darwin-arm64/
│   ├── wt-linux-x64/
│   ├── wt-linux-arm64/
│   └── wt-windows-x64/
├── scripts/
│   ├── publish-npm.sh         # Called by goreleaser
│   └── generate-wrappers.sh   # Generate shell wrapper templates
├── .goreleaser.yml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## Build Workflow

### Local Development
```bash
make build          # Build for current platform
make test           # Run tests
make lint           # Run linters
make install        # Install to $GOPATH/bin
```

### Release
```bash
# Tag release
git tag v0.1.0
git push origin v0.1.0

# goreleaser builds all platforms, creates GitHub release, publishes to npm
goreleaser release --clean

# Or test locally without publishing
goreleaser release --snapshot --clean
```

## Platform Support

| Platform | GOOS | GOARCH | npm Package | Priority |
|----------|------|--------|-------------|----------|
| macOS Intel | darwin | amd64 | @scope/wt-darwin-x64 | HIGH |
| macOS Apple Silicon | darwin | arm64 | @scope/wt-darwin-arm64 | HIGH |
| Linux x64 | linux | amd64 | @scope/wt-linux-x64 | HIGH |
| Linux ARM64 | linux | arm64 | @scope/wt-linux-arm64 | MEDIUM |
| Windows x64 | windows | amd64 | @scope/wt-windows-x64 | MEDIUM |
| Linux ARM | linux | arm | @scope/wt-linux-arm | LOW |
| Windows ARM | windows | arm64 | @scope/wt-windows-arm64 | LOW |

**Start with:** darwin/amd64, darwin/arm64, linux/amd64
**Add later:** linux/arm64, windows/amd64

## Confidence Assessment

| Component | Confidence | Notes |
|-----------|------------|-------|
| cobra | HIGH | De facto standard, verified usage in kubectl, gh, hugo |
| goreleaser | HIGH | Industry standard for Go binary distribution |
| npm pattern | HIGH | Proven by esbuild, biome, turbo (100M+ weekly downloads) |
| Versions | MEDIUM | Based on training data (Jan 2025); should verify latest stable |
| Shell wrapper approach | HIGH | Standard pattern for directory-changing CLIs |
| git binary vs go-git | HIGH | Common practice; simpler and more reliable |

## Version Verification Needed

**IMPORTANT:** Versions listed are based on training data through January 2025. Before implementation, verify:

1. cobra latest stable: `go list -m -versions github.com/spf13/cobra`
2. goreleaser latest: Check https://github.com/goreleaser/goreleaser/releases
3. golangci-lint latest: Check https://github.com/golangci/golangci-lint/releases
4. Go version: Check https://go.dev/dl/ (prefer latest stable)

## Sources

- cobra: https://github.com/spf13/cobra
- goreleaser: https://goreleaser.com/
- esbuild npm pattern: https://www.npmjs.com/package/esbuild (reference implementation)
- biome npm pattern: https://www.npmjs.com/package/@biomejs/biome (reference implementation)
- npm optionalDependencies: https://docs.npmjs.com/cli/v10/configuring-npm/package-json#optionaldependencies

## Anti-Patterns to Avoid

### 1. Single Fat Package
**Don't:** Create one npm package containing binaries for all platforms
**Why:** Users would download 50MB+ when they only need 5MB for their platform
**Instead:** Use optionalDependencies pattern (esbuild style)

### 2. postinstall Download Script
**Don't:** Use postinstall to download binary from GitHub releases
**Why:** Blocked by corporate firewalls, unreliable in CI, security concerns
**Instead:** Distribute binaries directly in npm packages via optionalDependencies

### 3. CGO Dependencies
**Don't:** Use libraries that require CGO (like git2go)
**Why:** Complicates cross-compilation, requires C compiler on user's machine
**Instead:** Shell out to git binary or use pure Go libraries (go-git)

### 4. Embedding Shell Wrappers in Binary
**Don't:** Have Go binary write shell wrapper to disk automatically
**Why:** Requires write permissions, unclear where to write, unexpected side effects
**Instead:** Interactive installer (npx @scope/wt install) prompts user and handles installation

### 5. Eval-Based Directory Changing
**Don't:** Use `eval $(wt goto main)` pattern
**Why:** Security risk, poor UX, requires users to remember syntax
**Instead:** Shell wrapper function that sources output and handles cd

### 6. Monolithic Command File
**Don't:** Put all cobra commands in one main.go file
**Why:** Hard to maintain, test, and review
**Instead:** One file per command in internal/commands/

## Next Steps for Implementation

1. Initialize Go module with cobra
2. Implement core commands (start with read-only: list)
3. Set up goreleaser for local testing
4. Create npm package structure
5. Implement interactive installer (install.js)
6. Test on all target platforms
7. Set up GitHub Actions for automated releases
