# Domain Pitfalls: Shell-to-Go CLI Rewrite with npm Distribution

**Domain:** Git worktree manager CLI (zsh to Go rewrite)
**Researched:** 2026-02-07
**Confidence:** HIGH (based on domain expertise for shell-to-binary migrations and cross-shell compatibility)

## Critical Pitfalls

Mistakes that cause rewrites or major issues.

### Pitfall 1: Subprocess Can't Change Parent Shell's Directory

**What goes wrong:** The Go binary cannot `cd` the parent shell. Commands like `wt goto <branch>` that need to change the user's working directory fail because the subprocess changes its own directory, not the parent shell's.

**Why it happens:** Operating system process isolation - child processes cannot modify parent process state (including `$PWD`).

**Consequences:**
- Core functionality (goto, home, new, eject) completely broken
- Users still in wrong directory after command execution
- Requires post-release architectural change (shell wrapper pattern)

**Prevention:**
1. Design from day one with shell wrapper pattern
2. Binary outputs `cd` command to stdout
3. Shell wrapper function captures output and evaluates it
4. Document this in architecture phase before implementation

**Detection:**
- Requirements include "commands that change directory"
- Testing reveals `pwd` unchanged after command
- Users report "nothing happens" or "I'm in the wrong directory"

**Phase impact:** Must be addressed in Phase 1 (Architecture) - affects all 4 cd-requiring commands.

---

### Pitfall 2: Shell Wrapper Syntax Incompatibility Across Shells

**What goes wrong:** Writing shell wrappers that work in bash 3.2, modern bash, zsh, and fish simultaneously is harder than expected. Each has different:
- Function definition syntax
- Variable scoping rules
- Command substitution behavior
- Array handling
- Conditional syntax

**Why it happens:**
- Bash 3.2 (macOS default) is from 2006, missing modern features
- Fish uses completely different syntax (not POSIX)
- Zsh has different expansion rules than bash
- Developer tests only in their primary shell

**Consequences:**
- Wrapper breaks in untested shells
- Silent failures (command appears to work but doesn't)
- `$()` vs `` ` ` `` differences cause parse errors
- `local` keyword behaves differently (bash vs zsh scoping)

**Prevention:**
1. Write separate wrapper files for each shell family:
   - `wt-wrapper.bash` (bash 3.2 compatible)
   - `wt-wrapper.zsh` (zsh specific)
   - `wt-wrapper.fish` (fish specific)
2. Test matrix: bash 3.2 (macOS), bash 5+ (Linux), zsh 5.8+, fish 3.0+
3. Use POSIX-compatible features only in bash wrapper (avoid bashisms)
4. Never use: `local -n`, `[[  ]]` in bash 3.2, arrays, `source` (use `.` in bash)
5. Fish wrapper is completely different - no shared code

**Detection:**
- Syntax errors when sourcing wrapper in different shell
- `command not found` errors
- Wrapper works in dev shell but not user's shell
- Bug reports from macOS users (bash 3.2) or fish users

**Phase impact:** Phase 2 (Shell Integration) - requires shell-specific implementations and thorough testing.

---

### Pitfall 3: npm Binary Distribution Platform Detection Failures

**What goes wrong:** Platform-specific binary selection fails due to:
- Inconsistent platform naming (`darwin` vs `macos`, `x64` vs `amd64`)
- Missing platform/arch combinations
- npm postinstall script execution failures
- Binary permissions not preserved

**Why it happens:**
- Node.js uses `process.platform` and `process.arch` with specific values
- Go builds use different naming (`GOOS`/`GOARCH`)
- npm doesn't preserve +x bit by default
- Platform detection code has untested edge cases

**Consequences:**
- Users get "binary not found" after install
- Wrong binary downloaded (x64 instead of arm64)
- Permission denied errors when executing binary
- Fails on less common platforms (Windows, arm64, Linux arm)

**Prevention:**
1. Use exact platform naming mapping:
   ```
   Node platform → GOOS/GOARCH
   darwin-x64    → darwin-amd64
   darwin-arm64  → darwin-arm64
   linux-x64     → linux-amd64
   linux-arm64   → linux-arm64
   win32-x64     → windows-amd64
   ```
2. Package structure: `@scope/wt-{platform}-{arch}` per platform
3. Main package `@scope/wt` has postinstall that:
   - Detects platform with `process.platform` and `process.arch`
   - Requires correct platform package as optionalDependency
   - Symlinks binary to bin/ directory
   - Runs `chmod +x` explicitly
4. Test on all target platforms before release
5. Document supported platforms in README

**Detection:**
- Installation succeeds but binary missing/non-executable
- `ENOENT` errors when running command
- Works on developer's platform but not others
- GitHub issues with "not working on M1 Mac" or "Linux fails"

**Phase impact:** Phase 3 (npm Distribution) - requires careful package.json setup and postinstall scripting.

---

### Pitfall 4: Shell RC File Modification Corrupts Existing Configuration

**What goes wrong:** Interactive installer modifies shell rc files (`.bashrc`, `.zshrc`, `.config/fish/config.fish`) but:
- Adds duplicate entries on re-run
- Breaks existing configuration syntax
- Doesn't handle sourced files (`.bash_profile` vs `.bashrc`)
- Corrupts file if write interrupted
- Doesn't respect user's existing wt installation

**Why it happens:**
- Installer appends without checking for existing entries
- No atomic write (write directly to rc file)
- Doesn't understand shell rc file sourcing chains
- Edge cases not tested (partial installs, interrupted writes)

**Consequences:**
- User's shell breaks (syntax errors on startup)
- Duplicate PATH entries slow shell startup
- User loses custom configuration if installer overwrites
- Uninstall impossible (installer code scattered throughout rc file)

**Prevention:**
1. **Idempotent installation:** Check for existing installation marker before adding
2. **Wrapped block pattern:**
   ```bash
   # >>> wt initialization >>>
   [installer code here]
   # <<< wt initialization <<<
   ```
3. **Atomic writes:** Write to temp file, validate, then move to target
4. **Backup before modify:** Copy original rc file to `.bashrc.wt-backup`
5. **Uninstall support:** Remove entire block between markers
6. **Shell detection:** Different logic for:
   - macOS bash (`.bash_profile` sourced, not `.bashrc`)
   - Linux bash (`.bashrc` sourced)
   - zsh (`.zshrc`)
   - fish (`config.fish`)
7. **User confirmation:** Show diff before applying changes

**Detection:**
- Shell startup errors after installation
- Multiple identical lines in rc file
- "command not found" for previously working commands
- User complaints about broken shell config

**Phase impact:** Phase 4 (Interactive Installer) - requires careful file manipulation and extensive testing.

---

### Pitfall 5: Losing Shell Script Semantics When Porting to Go

**What goes wrong:** Go's explicit error handling and type system don't map cleanly to shell script semantics:
- Shell: `cd /tmp || exit 1` (implicit error propagation)
- Go: Every call needs `if err != nil { return err }`
- Shell's loose string handling vs Go's types
- Exit codes and signal handling differences
- Environment variable expansion differences

**Why it happens:**
- Direct line-by-line translation without rethinking
- Not leveraging Go's advantages (types, testing, performance)
- Over-complicating simple shell operations
- Under-estimating shell's built-in features

**Consequences:**
- Go code is 3-5x longer than shell equivalent
- Error handling is inconsistent
- Subtle behavior changes break user workflows
- Miss shell features (parameter expansion, globs, pipes)

**Prevention:**
1. **Don't translate line-by-line** - understand intent, rewrite idiomatically
2. **Use stdlib effectively:**
   - `os/exec` for git commands
   - `filepath` for path manipulation
   - `os` for environment/filesystem
3. **Centralize error handling patterns:**
   ```go
   func mustExec(cmd *exec.Cmd) string {
       out, err := cmd.Output()
       if err != nil {
           log.Fatal(err)
       }
       return string(out)
   }
   ```
4. **Preserve shell semantics where it matters:**
   - Exit codes (use `os.Exit()` with correct codes)
   - Signal handling (SIGINT, SIGTERM)
   - Stdout/stderr separation
5. **Test against original shell script:** Same inputs should produce same outputs

**Detection:**
- Go code feels verbose and repetitive
- Behaviors differ from shell version
- Users report "it used to work differently"
- Error messages are less helpful than shell version

**Phase impact:** Phase 1 (Architecture & Design) - establish patterns before coding begins.

---

## Moderate Pitfalls

Mistakes that cause delays or technical debt.

### Pitfall 6: Cobra Completion Generation for Multiple Shells

**What goes wrong:** Cobra generates shell completions, but:
- Generated code doesn't work with wrapper functions
- Completions complete binary subcommands, not wrapper context
- Different quality across shells (bash vs zsh vs fish)
- Installation of completions varies by shell
- Completion scripts need manual tweaking

**Why it happens:**
- Cobra assumes direct binary execution
- Wrapper function adds indirection layer
- Completion script paths differ by OS and shell

**Prevention:**
1. Generate completions for binary, test in isolation
2. Wrapper function should pass through to binary for completion
3. Document completion installation per shell:
   - Bash: `complete -F _wt_completion wt`
   - Zsh: fpath addition + compinit
   - Fish: `wt completion fish | source`
4. Installer should handle completion installation
5. Test completions with wrapper installed

**Detection:**
- Tab completion doesn't work after installation
- Completions work for binary but not wrapper
- Shell-specific completion failures

**Phase impact:** Phase 2 (Shell Integration) - coordinate wrapper and completion together.

---

### Pitfall 7: Go Binary Size and Startup Time

**What goes wrong:** Go binaries are larger and slower to start than shell scripts:
- Static binary: 5-10 MB (vs 550-byte shell script)
- Cold start: 10-50ms (vs <1ms for shell)
- npm package: 50+ MB with all platform binaries

**Why it happens:**
- Go compiles stdlib into binary
- Static linking for portability
- No lazy loading like interpreted shells

**Prevention:**
1. Accept the tradeoff (correctness > size)
2. Optimization techniques:
   - `go build -ldflags="-s -w"` (strip debug info)
   - UPX compression (reduces 50-70% but slower start)
3. Don't optimize prematurely
4. npm platform packages (only download target platform)

**Detection:**
- Users complain about download size
- Slow performance on repeated calls
- Disk space issues

**Phase impact:** Phase 3 (npm Distribution) - use platform-specific packages to minimize download.

---

### Pitfall 8: Environment Variable Handling Differences

**What goes wrong:** Shell scripts inherit and modify environment freely. Go requires explicit handling:
- `export VAR=value` vs `os.Setenv()`
- Variable expansion: `$HOME` vs `os.ExpandEnv()`
- Subshell environment isolation
- Modified env doesn't persist to parent shell

**Why it happens:**
- Shell and Go have different process models
- Shell variable expansion is automatic
- Go requires explicit environment access

**Prevention:**
1. Use `os.Getenv()` for reading environment
2. For output that needs env vars: let shell wrapper handle expansion
3. Document environment requirements
4. Use `os.ExpandEnv()` for path expansion
5. Don't try to modify parent shell's environment (use shell wrapper)

**Detection:**
- `$HOME` appears literally in output instead of expanded path
- Environment variables not accessible
- Changes to env don't persist

**Phase impact:** Phase 1 (Architecture) - design env handling strategy early.

---

### Pitfall 9: Git Command Error Handling

**What goes wrong:** Shell scripts often ignore git command failures or handle them inconsistently. Go makes errors explicit but requires proper handling:
- `git` not in PATH
- Git repository not initialized
- Git commands fail with cryptic errors
- Different git versions have different behaviors

**Why it happens:**
- Shell: `git command || echo "failed"` is casual
- Go: Must check every `cmd.Run()` error
- Git error messages go to stderr, easy to lose

**Prevention:**
1. Wrapper helper for git commands:
   ```go
   func gitCmd(args ...string) (string, error) {
       cmd := exec.Command("git", args...)
       cmd.Stderr = os.Stderr
       out, err := cmd.Output()
       if err != nil {
           return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
       }
       return strings.TrimSpace(string(out)), nil
   }
   ```
2. Check for git availability at startup
3. Validate repository before operations
4. Preserve git's stderr output for user
5. Add context to error messages

**Detection:**
- Cryptic "exit status 1" errors
- Lost git error messages
- Silent failures
- No validation of git presence

**Phase impact:** Phase 1 (Architecture) - establish git interaction patterns.

---

### Pitfall 10: Testing Interactive CLI Without Automation

**What goes wrong:** Interactive prompts in installer are hard to test:
- Manual testing only
- No CI validation
- Regressions caught late
- Different terminal behavior (color, width)

**Why it happens:**
- Interactive code paths not designed for testing
- No stdin injection during tests
- Terminal detection hard to mock

**Prevention:**
1. Design for testability from start:
   ```go
   type Installer struct {
       In  io.Reader
       Out io.Writer
       Interactive bool
   }
   ```
2. Mock stdin/stdout in tests
3. Separate "prompt logic" from "IO":
   ```go
   func (i *Installer) Confirm(msg string) bool {
       if !i.Interactive {
           return false
       }
       // prompt on i.Out, read from i.In
   }
   ```
4. Add non-interactive mode: `--yes` flag
5. Test both interactive and non-interactive paths

**Detection:**
- Only manual testing possible
- CI can't test installer
- Regressions in interactive flows
- No test coverage for prompts

**Phase impact:** Phase 4 (Interactive Installer) - design for testability upfront.

---

## Minor Pitfalls

Mistakes that cause annoyance but are fixable.

### Pitfall 11: Hardcoded Paths Break Across Platforms

**What goes wrong:** Linux/macOS paths (`/usr/local`, `/home`) don't work on Windows.

**Prevention:**
- Use `os.UserHomeDir()` instead of `$HOME` or `/home/user`
- Use `filepath.Join()` for path construction
- Use `os.PathSeparator` instead of hardcoded `/`
- Test on Windows if supporting it

**Detection:** Windows users report "file not found" errors.

**Phase impact:** Phase 1 (Architecture) - use stdlib path functions.

---

### Pitfall 12: Color Output Breaks in Non-TTY Contexts

**What goes wrong:** ANSI color codes appear as garbage when output is piped or redirected.

**Prevention:**
- Detect TTY: `term.IsTerminal(int(os.Stdout.Fd()))`
- Disable colors for non-TTY
- Provide `--no-color` flag
- Respect `NO_COLOR` environment variable

**Detection:** Piped output shows `^[[31m` sequences.

**Phase impact:** Phase 5 (Polish) - add TTY detection.

---

### Pitfall 13: Version Information Missing or Incorrect

**What goes wrong:** Binary doesn't report version, or version is wrong/outdated.

**Prevention:**
- Use `-ldflags` to inject version at build time:
  ```bash
  go build -ldflags="-X main.version=$(git describe --tags)"
  ```
- Add `version` subcommand in cobra
- Include version in `--help` output

**Detection:** Users can't report what version they're using.

**Phase impact:** Phase 1 (Architecture) - set up version injection.

---

### Pitfall 14: Completion Script Installation Path Confusion

**What goes wrong:** Shell completion scripts need to be installed in shell-specific locations:
- Bash: `/usr/local/etc/bash_completion.d/` (macOS) or `/etc/bash_completion.d/` (Linux)
- Zsh: Anywhere in `$fpath`
- Fish: `~/.config/fish/completions/`

**Prevention:**
- Document manual installation per shell
- Installer detects shell and installs to correct location
- Don't assume standard paths exist
- Provide `wt completion <shell>` command to output script

**Detection:** Completions installed but not working.

**Phase impact:** Phase 4 (Interactive Installer) - handle per-shell installation.

---

### Pitfall 15: npm Package Naming and Scoping Mistakes

**What goes wrong:**
- Package name conflicts with existing packages
- Scoped package requires login to install
- Binary name doesn't match package name
- Platform packages not listed as optionalDependencies

**Prevention:**
- Use scoped package: `@username/wt`
- Main package `@username/wt` depends on platform packages as optionalDependencies
- Binary name in bin/ matches command name: `"wt": "./bin/wt"`
- Test installation before publishing

**Detection:**
- `npm install` fails with "package not found"
- Binary not in PATH after install
- Installation requires npm login

**Phase impact:** Phase 3 (npm Distribution) - get package.json structure right.

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| Architecture & Design | Subprocess can't cd parent shell | Design shell wrapper pattern from start |
| Architecture & Design | Losing shell script semantics | Don't translate line-by-line; rethink in Go idioms |
| Architecture & Design | Environment variable handling | Design env access strategy early |
| Shell Integration | Bash 3.2 vs zsh vs fish syntax | Separate wrapper files per shell; test matrix |
| Shell Integration | Cobra completions with wrapper | Coordinate wrapper and completion together |
| npm Distribution | Platform detection failures | Map Node/Go platform names; test all platforms |
| npm Distribution | Binary permissions not preserved | Explicit chmod +x in postinstall |
| Interactive Installer | RC file corruption | Atomic writes, backup, idempotent blocks |
| Interactive Installer | Duplicate entries on re-run | Check for markers before adding |
| Interactive Installer | Testing interactive prompts | Design for testability (mock IO) |
| Implementation | Git command error handling | Wrapper helper; preserve stderr |
| Implementation | Hardcoded paths | Use filepath stdlib |
| Polish | Color in non-TTY | Detect TTY; respect NO_COLOR |

---

## Shell Compatibility Gotchas

### Bash 3.2 (macOS Default) Limitations

- No associative arrays (only indexed arrays)
- No `local -n` (nameref)
- No `&>>` redirect shorthand
- No `[[  ]]` regex matching with `=~`
- `.bash_profile` sourced (not `.bashrc`) for login shells

**Prevention:** Use POSIX features only; test on macOS.

---

### Zsh Differences from Bash

- Arrays are 1-indexed (not 0-indexed)
- `setopt` changes behavior (word splitting, glob behavior)
- Different prompt expansion
- `local` has function scope (not block scope like bash)
- Special variable names (`path` is tied to `$PATH` - see MEMORY.md)

**Prevention:** Use separate zsh wrapper; avoid reserved names like `path`.

---

### Fish is Completely Different

- Not POSIX compatible
- Different function syntax: `function name; body; end`
- No `$( )` - use `( )` for command substitution
- Different variable syntax: `set var value`
- Different conditionals: `if test condition; ... end`

**Prevention:** Write separate fish wrapper from scratch; don't try to adapt bash/zsh version.

---

## Summary of Critical Prevention Strategies

1. **Shell wrapper pattern from day one** - don't discover need for cd post-launch
2. **Separate wrapper files per shell** - don't try to write one wrapper for all
3. **Test matrix: bash 3.2, bash 5, zsh, fish** - on both macOS and Linux
4. **npm platform packages with optionalDependencies** - correct platform detection
5. **RC file modification: idempotent blocks, atomic writes, backups** - don't corrupt user config
6. **Design for testability** - mock IO for interactive components
7. **Don't translate shell line-by-line** - rethink in Go idioms

---

## Confidence Notes

**HIGH confidence areas:**
- Shell wrapper pattern and subprocess limitations (fundamental OS constraint)
- Cross-shell syntax differences (well-documented, stable)
- npm binary distribution patterns (established ecosystem practice)
- Shell RC file modification risks (common pain point)

**MEDIUM confidence areas:**
- Cobra completion generation specifics (version-dependent)
- npm postinstall script best practices (evolving)

**Verification sources:**
- Shell wrapper pattern: OS process model (parent/child isolation)
- Bash 3.2 limitations: Official bash documentation and macOS defaults
- npm binary distribution: npm documentation and ecosystem patterns (e.g., esbuild, swc)
- Shell syntax differences: Official shell documentation (bash, zsh, fish)

**Note on research limitations:** Web search was unavailable, so this research draws on domain expertise and training knowledge. For current cobra/npm specifics, verify against:
- Cobra documentation: https://cobra.dev
- npm documentation: https://docs.npmjs.com
- Go wiki on binary distribution: https://github.com/golang/go/wiki

---

**Research completed:** 2026-02-07
**Researcher confidence:** HIGH for architectural and shell compatibility pitfalls; MEDIUM for tooling specifics
**Recommended next steps:** Verify cobra completion and npm packaging patterns with official documentation during implementation phases.
