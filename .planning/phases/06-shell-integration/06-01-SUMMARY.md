---
phase: 06-shell-integration
plan: 01
subsystem: shell-integration
tags: [go, shell, bash, zsh, fish, wrappers, eval, cobra]
requires:
  - 05-01-directory-changing-commands
  - 05-02-wt-new-command
  - 05-03-wt-eject-command
provides:
  - shell-detection-package
  - embedded-wrapper-scripts
  - shell-init-command
affects:
  - 06-02-shell-wrapper-documentation
  - 07-01-npm-package-setup
tech-stack:
  added: []
  patterns:
    - go:embed for shell wrapper templates
    - POSIX-compatible bash/zsh wrappers
    - Fish-specific syntax patterns
    - Hidden cobra commands for plumbing
key-files:
  created:
    - internal/shell/detect.go
    - internal/shell/detect_test.go
    - internal/shell/embed.go
    - internal/shell/templates/wrapper.bash
    - internal/shell/templates/wrapper.zsh
    - internal/shell/templates/wrapper.fish
    - cmd/shell_init.go
  modified: []
key-decisions:
  - decision: "POSIX-compatible bash wrapper"
    rationale: "Use [ ] not [[ ]] for bash 3.2+ compatibility (macOS default)"
    phase: "06-01"
  - decision: "Identical bash and zsh wrappers"
    rationale: "POSIX constructs work in both shells - no need for shell-specific features"
    phase: "06-01"
  - decision: "Hidden shell-init command"
    rationale: "Plumbing command for rc files, not user-facing - follows git-style UX"
    phase: "06-01"
  - decision: "Route all commands through wrapper"
    rationale: "Simpler mental model (like zoxide) - non-cd commands produce no output with --output-path"
    phase: "06-01"
duration: 2 min
completed: 2026-02-07
---

# Phase 06 Plan 01: Shell Wrapper Infrastructure Summary

**One-liner:** Shell detection, embedded bash/zsh/fish wrappers, and `wt shell-init` command for `eval $(wt shell-init)` rc file integration.

## Performance

**Duration:** 2 minutes
**Tasks completed:** 2/2
**Commits:** 2 (atomic per task)
**Tests added:** 5 unit tests for shell detection

## Accomplishments

Created the core shell integration infrastructure that enables `wt` to change the user's shell directory. Users add `eval $(wt shell-init)` to their rc file, which auto-detects the shell and outputs the appropriate wrapper function. All directory-changing commands (goto, home, new, eject, merge, rebase) now work seamlessly via the `--output-path` protocol established in Phase 5.

**Key deliverables:**

1. **Shell detection package** (`internal/shell/detect.go`) - Auto-detects bash/zsh/fish from $SHELL env var, returns error for unsupported shells
2. **Three embedded wrapper scripts** - POSIX-compatible bash/zsh wrappers and fish-specific wrapper, all using `command wt --output-path` bypass pattern
3. **wt shell-init command** - Hidden cobra command that outputs shell-specific wrapper, no arguments required

**Integration pattern:**
- User adds `eval $(wt shell-init)` to ~/.bashrc, ~/.zshrc, or ~/.config/fish/config.fish
- shell-init auto-detects shell type from $SHELL
- Outputs wrapper function to stdout (for eval capture)
- Wrapper routes ALL wt commands through function, performs cd only when command outputs a path

## Task Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1 | 77a4073 | feat(06-01): shell detection, embedded wrappers, and GetWrapper |
| 2 | ced8615 | feat(06-01): wt shell-init command |

## Files Created

**Internal shell package:**
- `internal/shell/detect.go` - DetectShell() function extracts shell name from $SHELL env var
- `internal/shell/detect_test.go` - Unit tests for bash/zsh/fish/csh/unset scenarios
- `internal/shell/embed.go` - go:embed directives and GetWrapper() retrieval function

**Embedded wrapper templates:**
- `internal/shell/templates/wrapper.bash` - Bash 3.2+ compatible wrapper (uses [ ] not [[ ]])
- `internal/shell/templates/wrapper.zsh` - Zsh wrapper (identical to bash for maintainability)
- `internal/shell/templates/wrapper.fish` - Fish wrapper with fish-specific syntax (set -l, $argv, $status)

**Command:**
- `cmd/shell_init.go` - wt shell-init cobra command (hidden from help)

## Files Modified

None - all new files.

## Decisions Made

**1. POSIX-compatible bash wrapper (bash 3.2+ on macOS)**
- **Context:** macOS ships with bash 3.2 (released 2006) due to GPL licensing
- **Decision:** Use `[ ]` not `[[ ]]`, `=` not `==`, avoid associative arrays
- **Rationale:** Ensures wrapper works on macOS default bash without requiring Homebrew upgrade
- **Impact:** Slightly more verbose syntax but maximum compatibility

**2. Identical bash and zsh wrappers**
- **Context:** Zsh supports many bash-isms and POSIX constructs
- **Decision:** Keep bash and zsh wrappers identical for maintainability
- **Rationale:** POSIX-compatible constructs work in both shells, no need for zsh-specific features
- **Impact:** Easier to maintain (one template to update), reduced testing surface
- **Note:** Avoided `path` variable per project memory (clobbers PATH in zsh)

**3. Hidden shell-init command**
- **Context:** shell-init is plumbing for rc files, not a user-facing command
- **Decision:** Set `Hidden: true` in cobra command definition
- **Rationale:** Follows git-style UX (plumbing vs porcelain commands)
- **Impact:** Cleaner help output, power users can still discover via tab completion

**4. Route all commands through wrapper**
- **Context:** Could filter to only wrap cd-producing commands (goto, home, new, eject)
- **Decision:** Route ALL wt commands through the wrapper function
- **Rationale:** Simpler mental model (like zoxide) - non-cd commands produce no output with --output-path flag, so wrapper is a pass-through
- **Impact:** Consistent user experience, no need to remember which commands need the wrapper

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - all tests passed, binary compiled successfully, verifications confirmed functionality.

## Next Phase Readiness

**Blockers:** None

**Recommendations:**
1. Document the shell integration in user-facing docs (06-02)
2. Add shell-init to installation instructions
3. Consider adding shell-init output example to README

**Dependencies satisfied:**
- Phase 5 (Directory-Changing Commands) ✓ Complete - all cd-producing commands use --output-path protocol
- Internal git package ✓ Available - worktree path resolution works
- Cobra framework ✓ Integrated - hidden commands supported

**Ready for:**
- Phase 06 Plan 02: Shell wrapper documentation and installation instructions
- Phase 07: npm package setup and distribution

## Self-Check: PASSED

All files from key-files.created exist on disk:
- FOUND: internal/shell/detect.go
- FOUND: internal/shell/detect_test.go

All commits exist in git history:
- 77a4073 feat(06-01): shell detection, embedded wrappers, and GetWrapper
- ced8615 feat(06-01): wt shell-init command
