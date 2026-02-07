# Project Research Summary

**Project:** wt (Git Worktree Manager - Go Rewrite)
**Domain:** CLI tool with npm distribution and shell integration
**Researched:** 2026-02-07
**Confidence:** HIGH

## Executive Summary

The wt project is a Git worktree manager being rewritten from a 550-line zsh script to a Go binary with multi-shell support (bash/zsh/fish) distributed via npm. Expert practice for this domain follows a well-established three-layer architecture: Go binary for business logic, thin shell wrappers for directory-changing commands (cd cannot be executed by subprocesses), and npm distribution using platform-specific optional dependencies (the esbuild/turbo/biome pattern). This approach is proven at scale with 100M+ weekly downloads.

The recommended technical approach is straightforward: use cobra CLI framework for command structure and built-in completions, goreleaser for cross-platform builds and npm publishing, and shell out to the git binary (avoid go-git library complexity). The architecture separates 9 commands into two categories: 4 cd-requiring commands (new, goto, home, eject) that need shell wrapper coordination via `--output-path` flag, and 5 standard commands (init, list, merge, rebase, delete) that work as pure Go subprocesses.

The critical risk is subprocess directory-changing limitations combined with cross-shell syntax incompatibility. Mitigation requires designing the shell wrapper pattern from day one (not post-launch) and writing separate wrapper files for each shell (bash 3.2, zsh, fish have incompatible syntax). Secondary risks include npm platform detection failures, shell rc file corruption during installation, and losing shell script semantics when translating to Go. These are all preventable through established patterns: platform-specific optional dependencies, idempotent rc file modification with marker blocks, and git command abstraction layers.

## Key Findings

### Recommended Stack

The Go ecosystem provides proven tools for CLI binary distribution via npm. The core stack is minimal: Go 1.21+ with cobra for CLI structure (used by kubectl, gh, hugo), goreleaser for multi-platform builds and npm publishing, and the stdlib testing package. The npm distribution follows the esbuild pattern: main wrapper package with platform-specific optional dependencies, ensuring users only download binaries for their platform (5MB vs 50MB+ fat package).

**Core technologies:**
- **Go 1.21+ with cobra v1.8+**: CLI framework with built-in bash/zsh/fish/powershell completion generation — de facto standard used by kubectl, gh, hugo
- **goreleaser v1.23+**: Handles cross-compilation, npm publishing, checksums, and archives — industry standard for Go binary distribution
- **npm optional dependencies pattern**: Main package depends on platform packages (@scope/wt-darwin-arm64, etc.) — proven by esbuild (100M+ weekly downloads), biome, turbo

**Critical version requirements:**
- Go 1.21+ for generics and improved error handling (1.22+ preferred)
- Bash 3.2 compatibility for macOS (default shell is ancient but cannot require brew install bash)

**Anti-patterns to avoid:**
- Do NOT use go-git library (shell out to git binary instead — simpler, respects user's git config, smaller binary)
- Do NOT use postinstall download scripts (blocked by corporate firewalls, unreliable)
- Do NOT bundle all binaries in one npm package (50MB+ download vs 5MB needed)
- Do NOT use CGO dependencies (complicates cross-compilation)

### Expected Features

Port all 9 existing wt.zsh commands with exact feature parity — no additions, no removals. Modern CLI table stakes include shell completion generation, --help/--version flags, platform-specific binaries, proper exit codes, colored output with NO_COLOR support, and interactive installer. The differentiators are suffix-based worktree matching, automatic shell detection, override flags (--copy/--symlink), and smart eject with stash handling.

**Must have (table stakes):**
- Shell completion generation for bash/zsh/fish — cobra provides built-in completion subcommand
- Platform-specific binaries via npm — requires architecture-specific packages or optionalDependencies
- Interactive installer — npm postinstall that detects shell, offers to modify rc files with confirmation
- Shell wrapper generation — commands that cd (goto, home, new, eject) need shell functions sourced into user's shell
- Config file parsing (.wtconfig) — existing users depend on line-based `<action> <path>` format
- Worktree name resolution — suffix matching on directory basename (existing feature)

**Should have (competitive):**
- Automatic shell detection — zero-config installer detects from SHELL env var
- Override flags (--copy/--symlink) — per-command overrides for .wtconfig behavior
- Smart eject with stash handling — safely move current branch to new worktree with uncommitted changes
- Suffix-based worktree matching — quality-of-life feature differentiating from raw git worktree
- Single binary + shell wrapper — unique approach for npm distribution

**Defer (v2+):**
- Colored output — focus on correctness first, add color later
- Advanced error messages — iterate after core works
- Non-interactive mode flags — add when CI users request it
- Windows native support — WSL is sufficient for Git workflows

**Anti-features (explicitly NOT build):**
- Built-in cd implementation — cannot cd from child process, use shell wrapper
- Complex config file format (YAML/TOML) — keep simple line-based format for hand-editing
- Automatic rc file modification — prompt user, require confirmation, show manual fallback
- Git subcommand (git wt) — standalone command is simpler mental model
- Interactive TUI — command-focused, leverage shell history
- Auto-update mechanism — npm handles updates

### Architecture Approach

Three-layer architecture separates concerns cleanly: shell wrapper functions (sourced into user's shell) handle directory changes by calling the Go binary with `--output-path` flag, the Go binary contains all business logic and outputs structured responses (paths for cd, or human-readable messages), and the npm package uses platform-specific optional dependencies to minimize download size. This pattern is proven by nvm, zoxide, direnv for shell integration and esbuild, biome, turbo for npm binary distribution.

**Major components:**
1. **Shell Layer (bash/zsh/fish wrappers)** — thin functions sourced into shell, invoke Go binary for cd commands (new/goto/home/eject) with --output-path flag, capture stdout (path), execute cd in current shell context
2. **Go Binary Layer** — all git operations, validation, config parsing; outputs structured responses; no parent process side effects; cobra command tree with one file per command in internal/commands/
3. **npm Distribution Layer** — main package (@scope/wt) with postinstall script; platform-specific optional dependencies (only matching platform installs); interactive installer for shell detection and rc file modification
4. **Git Abstraction (internal/git/)** — wrapper functions around git CLI invocations; makes testing easier via interface mocking; preserves git stderr for user visibility

**Key patterns:**
- **Command categorization**: 4 cd commands (new/goto/home/eject) need --output-path flag and shell wrapper; 5 standard commands (init/list/merge/rebase/delete) work as pure Go binaries
- **Config override chain**: Merge .wtconfig entries with CLI flag overrides (--copy/--symlink flags override config entries for that command)
- **Structured output mode**: --output-path outputs just path for shell wrapper; default mode outputs human-readable colored messages
- **Cobra command factory functions**: Each command is function returning *cobra.Command with local flag variables (avoid global state)

### Critical Pitfalls

The top pitfalls are architectural (must be addressed in Phase 1 design) and shell compatibility (must handle in Phase 2 integration). All are preventable through established patterns.

1. **Subprocess can't change parent shell's directory** — Operating system process isolation prevents Go binary from changing user's $PWD. Design shell wrapper pattern from day one: binary outputs path via --output-path flag, shell wrapper captures stdout and executes cd. Affects all 4 cd commands (goto, home, new, eject). MUST be in architecture before implementation starts.

2. **Shell wrapper syntax incompatibility across shells** — Bash 3.2 (macOS default, from 2006) lacks modern features, zsh has different scoping rules (reserved variable name `path` tied to $PATH), fish is completely non-POSIX. Write separate wrapper files (wt.bash, wt.zsh, wt.fish) with no shared code. Test matrix: bash 3.2 (macOS), bash 5+ (Linux), zsh 5.8+, fish 3.0+. Never use: local -n, [[]], arrays in bash wrapper; avoid reserved names like `path` in zsh.

3. **npm binary distribution platform detection failures** — Inconsistent platform naming (darwin vs macos, x64 vs amd64) causes wrong binary download. Use exact mapping: darwin-x64 → darwin-amd64, darwin-arm64 → darwin-arm64, linux-x64 → linux-amd64. Package structure: @scope/wt-{platform}-{arch} per platform as optionalDependencies in main package. Postinstall must chmod +x explicitly (npm doesn't preserve permissions).

4. **Shell rc file modification corrupts existing configuration** — Installer appending to .bashrc/.zshrc without checking for existing entries causes duplicates, interrupted writes corrupt files, no uninstall path. Use idempotent marker blocks (# >>> wt initialization >>>, # <<< wt initialization <<<), atomic writes to temp file then move, backup original to .bashrc.wt-backup, prompt user before modifying, handle shell-specific rc files (macOS bash uses .bash_profile not .bashrc).

5. **Losing shell script semantics when porting to Go** — Line-by-line translation produces verbose Go (3-5x longer) with inconsistent error handling. Don't translate literally; rethink intent in Go idioms. Centralize git command wrappers, preserve shell semantics where it matters (exit codes, stdout/stderr separation, signal handling), test against original zsh script for output parity.

## Implications for Roadmap

Based on research, suggested phase structure follows dependency ordering: core Go binary without cd commands first (establishes patterns and git abstraction), config system next (used by cd commands), then complex cd commands (depends on output formatting), shell integration (depends on --output-path support), npm distribution (depends on compiled binaries), interactive installer (depends on shell wrappers), and finally polish.

### Phase 1: Core Go Binary Foundation
**Rationale:** Establish project structure, git abstraction, and simple commands before tackling cd complexity. Build confidence with commands that don't need shell wrapper coordination.

**Delivers:** Working wt binary with list, init, delete commands; cobra structure; git abstraction layer

**Addresses:**
- Git abstraction (internal/git/) for all git worktree/branch operations
- Simple commands (list, init, delete) that work as pure Go binaries
- Cobra command structure with factory functions (avoid global state pitfall)
- Version injection via ldflags

**Avoids:**
- Pitfall 5: Losing shell semantics — establish git command wrapper pattern early
- Pitfall 9: Git error handling — centralized git command abstraction with context-specific errors

**Research flag:** Standard patterns, skip research-phase

### Phase 2: Configuration System
**Rationale:** .wtconfig parsing and setup logic is used by cd commands (new, eject), so implement before Phase 3. Config override chain (CLI flags override .wtconfig entries) must be designed before implementation.

**Delivers:** .wtconfig parser, copy/symlink setup logic, override flags (--copy/--symlink)

**Addresses:**
- Parse line-based .wtconfig format (existing users depend on this)
- Apply copy/symlink actions when creating worktrees
- Override flags that merge with config entries (CLI flags take precedence)
- Config-free operation (works without .wtconfig)

**Uses:** Go stdlib (os, filepath, bufio for parsing)

**Avoids:**
- Anti-pattern: Complex config format — keep line-based `<action> <path>` for hand-editing
- Pitfall: Config validation edge cases (missing files, invalid actions)

**Research flag:** Standard patterns, skip research-phase

### Phase 3: Directory-Changing Commands
**Rationale:** Implement new, goto, home, eject commands with --output-path flag support. These are the core value proposition (effortless worktree navigation).

**Delivers:** All 4 cd commands outputting paths in machine-readable mode; resolver for worktree name matching

**Addresses:**
- Commands: new, goto, home, eject with --output-path flag
- Resolver for suffix-based worktree name matching
- Structured output formatter (--output-path vs human-readable)
- Smart eject with stash handling

**Uses:**
- Config system from Phase 2 (new/eject apply .wtconfig actions)
- Git abstraction from Phase 1

**Implements:** Binary output protocol (path-only mode for shell wrapper consumption)

**Avoids:**
- Pitfall 1: Subprocess can't cd — binary outputs path, shell wrapper will cd in Phase 4
- Pitfall 5: Eject complexity — port stash handling logic carefully

**Research flag:** Eject command is complex, may need deeper research for stash handling edge cases

### Phase 4: Shell Integration
**Rationale:** Depends on --output-path support from Phase 3. Write separate wrappers per shell (bash/zsh/fish) with no shared code.

**Delivers:** Shell wrapper functions (wt.bash, wt.zsh, wt.fish); shell completions

**Addresses:**
- Thin wrapper functions for each shell (separate files, no shared code)
- Wrapper detects cd commands (new/goto/home/eject), calls binary with --output-path
- Captures stdout (path), executes cd in current shell
- Completions for bash/zsh/fish (cobra generated, manual testing)

**Avoids:**
- Pitfall 2: Shell syntax incompatibility — separate files for bash 3.2, zsh, fish
- Pitfall 6: Cobra completions with wrapper — coordinate wrapper and completion together
- Anti-pattern: One wrapper for all shells — fish is completely different syntax

**Research flag:** CRITICAL PHASE — needs extensive cross-shell testing. Test matrix: bash 3.2 (macOS), bash 5+ (Linux), zsh 5.8+, fish 3.0+

### Phase 5: npm Distribution
**Rationale:** Depends on compiled binaries from Phases 1-3. Platform detection must be bulletproof.

**Delivers:** npm package structure with platform-specific optional dependencies; goreleaser config; build scripts

**Addresses:**
- Main package (@scope/wt) with platform-specific optional dependencies
- Platform packages: @scope/wt-darwin-arm64, @scope/wt-darwin-x64, @scope/wt-linux-x64, etc.
- goreleaser config for cross-compilation and npm publishing
- Postinstall script that copies binary and runs installer

**Uses:**
- goreleaser for builds (STACK.md recommendation)
- npm optional dependencies pattern (esbuild precedent)

**Avoids:**
- Pitfall 3: Platform detection failures — exact mapping of Node platform → Go GOOS/GOARCH
- Pitfall 7: Binary size — use platform packages (5MB download vs 50MB+ fat package)
- Anti-pattern: postinstall download script — distribute binaries in npm packages

**Research flag:** goreleaser npm publishing may need verification with official docs

### Phase 6: Interactive Installer
**Rationale:** Depends on shell wrappers from Phase 4. RC file modification must be idempotent and safe.

**Delivers:** Interactive installer that detects shell, prompts for confirmation, modifies rc files safely

**Addresses:**
- Shell detection from SHELL env var
- RC file location logic (bash: .bashrc or .bash_profile, zsh: .zshrc, fish: config.fish)
- Idempotent installation (check for markers, skip if already installed)
- User confirmation before modifying files
- Atomic writes with backup
- Manual fallback instructions

**Avoids:**
- Pitfall 4: RC file corruption — idempotent blocks, atomic writes, backups
- Pitfall 10: Testing interactive CLI — design for testability (mock IO)
- Anti-pattern: Silent postinstall — prompt user, show what will be added

**Research flag:** Edge cases need careful testing (interrupted writes, multiple rc files, already installed)

### Phase 7: Polish & Documentation
**Rationale:** After core functionality works, improve UX and error messages.

**Delivers:** Improved error messages, colored output with TTY detection, comprehensive documentation

**Addresses:**
- Error messages with actionable suggestions
- Colored output with TTY detection and NO_COLOR support
- README with installation, usage, examples
- Troubleshooting guide for common issues
- CI/CD for automated releases

**Avoids:**
- Pitfall 12: Color in non-TTY — detect TTY, disable colors for pipes
- Pitfall 13: Missing version info — version already injected in Phase 1

**Research flag:** Standard patterns, skip research-phase

### Phase Ordering Rationale

- **Phases 1-3 are sequential**: Config system depends on project structure, cd commands depend on config system
- **Phase 4 depends on Phase 3**: Shell wrappers need --output-path flag support
- **Phase 5 depends on Phases 1-4**: Need compiled binaries and wrappers to distribute
- **Phase 6 depends on Phase 4**: Installer needs wrappers to install
- **Phase 7 is parallel-friendly**: Polish can overlap with testing earlier phases

This order follows the critical path: Core → Config → CD Commands → Shell Integration → Distribution → Installer → Polish. It avoids Pitfall 1 (subprocess cd) by designing shell wrapper protocol in Phase 3-4 before implementation, and avoids Pitfall 2 (shell syntax) by separating wrappers in Phase 4 with extensive testing.

### Research Flags

**Phases needing deeper research during planning:**
- **Phase 3 (Eject command)**: Complex stash handling, branch fallback logic — may need research-phase for edge cases
- **Phase 4 (Shell integration)**: CRITICAL for cross-shell testing — bash 3.2 limitations, zsh reserved variables, fish syntax
- **Phase 5 (goreleaser npm)**: Verify goreleaser npm publishing configuration with official docs
- **Phase 6 (Installer edge cases)**: RC file modification edge cases, shell detection failures

**Phases with standard patterns (skip research-phase):**
- **Phase 1**: Standard Go CLI project setup, well-documented cobra patterns
- **Phase 2**: Simple config file parsing, established patterns
- **Phase 7**: Standard polish tasks, no novel patterns

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | cobra and goreleaser are industry standard; versions need verification (based on Jan 2025 training data) |
| Features | HIGH | Based on direct analysis of existing wt.zsh codebase |
| Architecture | HIGH | Three-layer pattern proven by nvm/zoxide (shell) and esbuild/biome (npm) |
| Pitfalls | HIGH | Subprocess cd limitation is fundamental OS constraint; shell syntax differences are well-documented |

**Overall confidence:** HIGH for architectural patterns and pitfall prevention; MEDIUM for specific tool versions (cobra v1.8+, goreleaser v1.23+ should be verified)

### Gaps to Address

Specific library API versions and npm packaging best practices may have evolved since training data (Jan 2025). Verify during implementation:

- **Cobra completion API**: Verify current syntax for custom completions (worktree name completion)
- **goreleaser npm publishing**: Confirm npm publisher configuration syntax with official docs
- **npm optionalDependencies behavior**: Verify platform matching logic with npm documentation
- **Fish shell wrapper syntax**: Test wrapper example on actual fish 3.0+ (syntax verified from docs but not runtime tested)

**Mitigation strategy**: For each phase with research flags, validate with official documentation (cobra.dev, goreleaser.com, docs.npmjs.com) before implementation. The architectural patterns are sound; specific APIs just need version verification.

## Sources

### Primary (HIGH confidence)
- wt.zsh codebase analysis (direct read) — feature requirements, existing command behavior
- OS process model (fundamental constraint) — subprocess directory-changing limitation
- Shell documentation (bash, zsh, fish official docs) — syntax incompatibilities, version limitations
- npm documentation and ecosystem patterns — optional dependencies pattern, esbuild/biome/turbo precedent

### Secondary (MEDIUM confidence)
- cobra CLI framework (training data, pre-2025) — command structure, completion generation
- goreleaser build tool (training data, pre-2025) — multi-platform builds, npm publishing
- Go project layout conventions (golang-standards/project-layout community standard)

### Tertiary (LOW confidence, needs validation)
- Specific cobra v1.8+ completion API — verify with cobra.dev
- goreleaser v1.23+ npm publisher syntax — verify with goreleaser.com
- npm optionalDependencies platform matching — verify with docs.npmjs.com
- Current best practices for NO_COLOR support — verify with no-color.org

---
*Research completed: 2026-02-07*
*Ready for roadmap: yes*
