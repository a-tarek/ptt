# Requirements: wt v2.0 Go Rewrite

**Defined:** 2026-02-07
**Core Value:** A single `wt` command that works in any shell on any platform with full autocompletion

## v2.0 Requirements

### Go Binary — Command Port

- [x] **CMD-01**: `wt new [--copy <path>] [--symlink <path>] <name> [branch]` creates worktree with optional config overrides
- [x] **CMD-02**: `wt goto <worktree>` outputs target path for shell wrapper cd
- [x] **CMD-03**: `wt home` outputs main worktree path for shell wrapper cd
- [x] **CMD-04**: `wt init` creates .wtconfig template with commented examples
- [x] **CMD-05**: `wt eject [--copy <path>] [--symlink <path>] [name]` ejects branch with stash handling
- [x] **CMD-06**: `wt list` displays all worktrees with current marker
- [x] **CMD-07**: `wt merge <worktree>` merges worktree branch into current
- [x] **CMD-08**: `wt rebase <worktree>` rebases current onto worktree branch
- [x] **CMD-09**: `wt delete <worktree>` removes worktree

### Go Binary — Infrastructure

- [x] **INFRA-01**: .wtconfig parsing (copy/symlink actions, comments, blank lines)
- [x] **INFRA-02**: Override flag merging (--copy/--symlink override .wtconfig per-path)
- [x] **INFRA-03**: Worktree name resolution (suffix matching on directory basename)
- [x] **INFRA-04**: `--help` and `--version` flags
- [x] **INFRA-05**: Proper exit codes (0=success, 1=error), stderr for errors

### Shell Integration

- [x] **SHELL-01**: CD directive protocol (binary outputs path via --output-path, wrapper handles cd)
- [x] **SHELL-02**: Bash wrapper function (bash 3.2 compatible)
- [x] **SHELL-03**: Zsh wrapper function
- [x] **SHELL-04**: Fish wrapper function

### Completions

- [x] **COMP-01**: `wt completion bash` generates bash completions
- [x] **COMP-02**: `wt completion zsh` generates zsh completions
- [x] **COMP-03**: `wt completion fish` generates fish completions
- [x] **COMP-04**: Dynamic worktree name completion for goto/merge/rebase/delete

### npm Distribution

- [x] **NPM-01**: Scoped npm package (@scope/wt) with platform-specific binaries
- [x] **NPM-02**: Platform detection (darwin-arm64, darwin-amd64, linux-amd64, linux-arm64)
- [x] **NPM-03**: goreleaser config for cross-compilation

### Interactive Installer

- [ ] **INST-01**: `npx @scope/wt install` detects shell from $SHELL
- [ ] **INST-02**: Shows what will be added to rc file, requires confirmation
- [ ] **INST-03**: Idempotent — detects existing installation
- [ ] **INST-04**: Provides manual instructions if user declines

## Future Requirements

### Post-v2.0

- **COLOR-01**: Colored output with NO_COLOR support
- **CI-01**: Non-interactive mode flags for CI/scripting
- **ERR-01**: Enhanced error messages with suggestions

## Out of Scope

| Feature | Reason |
|---------|--------|
| Claude Code skill | Claude runs git commands natively, minimal added value |
| PowerShell support | Focus on bash/zsh/fish; WSL covers Windows |
| New features beyond wt.zsh | Port only — clean port reduces risk |
| GUI installer | Terminal-based interactive installer is sufficient |
| Git subcommand (`git wt`) | Standalone command is simpler mental model |
| Auto-update mechanism | npm handles updates |
| Interactive TUI | Commands are one-liners, TUI adds complexity |
| go-git library | Shell out to git binary — simpler, respects user's git config |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| CMD-01 | Phase 5 | Complete |
| CMD-02 | Phase 5 | Complete |
| CMD-03 | Phase 5 | Complete |
| CMD-04 | Phase 3 | Complete |
| CMD-05 | Phase 5 | Complete |
| CMD-06 | Phase 3 | Complete |
| CMD-07 | Phase 5 | Complete |
| CMD-08 | Phase 5 | Complete |
| CMD-09 | Phase 3 | Complete |
| INFRA-01 | Phase 4 | Complete |
| INFRA-02 | Phase 4 | Complete |
| INFRA-03 | Phase 3 | Complete |
| INFRA-04 | Phase 3 | Complete |
| INFRA-05 | Phase 3 | Complete |
| SHELL-01 | Phase 6 | Complete |
| SHELL-02 | Phase 6 | Complete |
| SHELL-03 | Phase 6 | Complete |
| SHELL-04 | Phase 6 | Complete |
| COMP-01 | Phase 6 | Complete |
| COMP-02 | Phase 6 | Complete |
| COMP-03 | Phase 6 | Complete |
| COMP-04 | Phase 6 | Complete |
| NPM-01 | Phase 7 | Complete |
| NPM-02 | Phase 7 | Complete |
| NPM-03 | Phase 7 | Complete |
| INST-01 | Phase 8 | Pending |
| INST-02 | Phase 8 | Pending |
| INST-03 | Phase 8 | Pending |
| INST-04 | Phase 8 | Pending |

**Coverage:**
- v2.0 requirements: 28 total
- Mapped to phases: 28 (100% coverage)
- Unmapped: 0

---
*Requirements defined: 2026-02-07*
*Last updated: 2026-02-07 with phase mappings*
