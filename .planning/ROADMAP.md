# Roadmap: wt

## Milestones

- ✅ **v1.0 Documentation** - Phases 1-2 (shipped 2026-02-07)
- ✅ **v2.0 Go Rewrite** - Phases 3-9 (shipped 2026-02-08)

## Phases

<details>
<summary>✅ v1.0 Documentation (Phases 1-2) - SHIPPED 2026-02-07</summary>

### Phase 1: Internal Documentation Refresh
**Goal**: Accurate codebase map reflecting all current wt.zsh features and structure
**Plans**: 3 plans

Plans:
- [x] 01-01-PLAN.md — Update ARCHITECTURE.md and STRUCTURE.md with new functions and accurate line ranges
- [x] 01-02-PLAN.md — Update CONVENTIONS.md and CONCERNS.md with flag parsing patterns and current concerns
- [x] 01-03-PLAN.md — Update STACK.md, INTEGRATIONS.md, TESTING.md with current state

### Phase 2: User-Facing Documentation
**Goal**: Complete README.md enabling users to install, learn, and use every wt feature
**Plans**: 3 plans

Plans:
- [x] 02-01-PLAN.md — Create README.md with header, installation, quick start, wt init, and tab completion
- [x] 02-02-PLAN.md — Add wt new, wt eject, wt goto, wt home, wt list command reference
- [x] 02-03-PLAN.md — Add wt merge/rebase/delete, configuration section, and container workflow

</details>

<details>
<summary>✅ v2.0 Go Rewrite (Phases 3-9) - SHIPPED 2026-02-08</summary>

### ✅ v2.0 Go Rewrite (Complete)

**Milestone Goal:** Rewrite wt in Go for cross-platform, multi-shell support with npm distribution

#### Phase 3: Core Go Binary Foundation
**Goal**: Establish project structure with simple commands that don't require shell integration
**Depends on**: Nothing (first phase of v2.0)
**Requirements**: INFRA-03, INFRA-04, INFRA-05, CMD-06, CMD-04, CMD-09
**Success Criteria** (what must be TRUE):
  1. User can run `wt --version` and see version number
  2. User can run `wt list` and see all worktrees with current marker
  3. User can run `wt init` to create .wtconfig template
  4. User can run `wt delete <worktree>` to remove a worktree
  5. Binary returns proper exit codes (0=success, 1=error) with errors to stderr
**Plans**: 2 plans

Plans:
- [x] 03-01-PLAN.md — Go project scaffold with cobra, wt list, and wt init commands
- [x] 03-02-PLAN.md — Worktree name resolution and wt delete command

#### Phase 4: Configuration System
**Goal**: Parse .wtconfig and handle copy/symlink actions with CLI flag overrides
**Depends on**: Phase 3
**Requirements**: INFRA-01, INFRA-02
**Success Criteria** (what must be TRUE):
  1. User can create .wtconfig with copy/symlink actions
  2. Binary correctly parses .wtconfig (handles comments, blank lines)
  3. Override flags (--copy/--symlink) correctly merge with .wtconfig entries
  4. Binary works correctly without .wtconfig (config-free operation)
**Plans**: 2 plans

Plans:
- [x] 04-01-PLAN.md — Config parsing, validation, resolution, inline flags, and init command updates (TDD)
- [x] 04-02-PLAN.md — Setup action execution (copy/symlink/run) with rollback (TDD)

#### Phase 5: Directory-Changing Commands
**Goal**: Implement all cd commands with structured output for shell wrapper coordination
**Depends on**: Phase 4
**Requirements**: CMD-01, CMD-02, CMD-03, CMD-05, CMD-07, CMD-08
**Success Criteria** (what must be TRUE):
  1. Binary outputs target path when invoked with --output-path flag
  2. User can create new worktree with `wt new` (applying .wtconfig actions)
  3. User can resolve worktree by suffix match (suffix-based name resolution works)
  4. User can eject current branch with stash handling
  5. User can merge and rebase worktree branches
**Plans**: 3 plans

Plans:
- [x] 05-01-PLAN.md — Shared git helpers + goto, home, merge, rebase commands
- [x] 05-02-PLAN.md — wt new command with config integration
- [x] 05-03-PLAN.md — wt eject command with stash handling and rollback

#### Phase 6: Shell Integration
**Goal**: Provide shell wrapper functions and completions for bash, zsh, and fish
**Depends on**: Phase 5
**Requirements**: SHELL-01, SHELL-02, SHELL-03, SHELL-04, COMP-01, COMP-02, COMP-03, COMP-04
**Success Criteria** (what must be TRUE):
  1. User can source bash wrapper and `wt goto <worktree>` changes directory in current shell
  2. User can source zsh wrapper and cd commands work without breaking PATH lookups
  3. User can source fish wrapper and cd commands work
  4. User can tab-complete worktree names for goto/merge/rebase/delete
  5. User can generate completions for their shell with `wt completion <shell>`
**Plans**: 2 plans

Plans:
- [x] 06-01-PLAN.md — Shell wrapper infrastructure (detect, embed, bash/zsh/fish scripts) and wt shell-init command
- [x] 06-02-PLAN.md — Dynamic worktree name completions for goto/delete/merge/rebase and completion command

#### Phase 7: npm Distribution
**Goal**: Package Go binaries for multi-platform distribution via npm with automatic platform detection
**Depends on**: Phase 6
**Requirements**: NPM-01, NPM-02, NPM-03
**Success Criteria** (what must be TRUE):
  1. User can `npm install @scope/wt` and get correct binary for their platform
  2. Only the platform-specific binary is downloaded (not all platforms)
  3. goreleaser config produces npm-compatible packages for all supported platforms
  4. Binary is executable after npm install (chmod +x applied)
**Plans**: 2 plans

Plans:
- [x] 07-01-PLAN.md — GoReleaser cross-compilation config and npm package structure (main wrapper + 4 platform packages)
- [x] 07-02-PLAN.md — Build/publish automation scripts and local pipeline validation

#### Phase 8: Interactive Installer
**Goal**: Provide npx-based installer that safely modifies shell rc files with user confirmation
**Depends on**: Phase 7
**Requirements**: INST-01, INST-02, INST-03, INST-04
**Success Criteria** (what must be TRUE):
  1. User can run `npx @scope/wt install` and see detected shell
  2. Installer shows what will be added to rc file before modifying
  3. Running installer twice doesn't duplicate entries (idempotent)
  4. User declining installation gets manual instructions
  5. RC file modifications use marker blocks for safe uninstall
**Plans**: 2 plans

Plans:
- [x] 08-01-PLAN.md — Installer package (RC file operations, marker blocks, v1 migration) and wt install command with guided walkthrough
- [x] 08-02-PLAN.md — wt uninstall command with confirmation and rc file cleanup

#### Phase 9: Polish & Testing
**Goal**: Improve error messages with actionable suggestions, add CI/CD pipeline, comprehensive shell testing, and rewrite README for v2
**Depends on**: Phase 8
**Requirements**: (No specific requirements - polish existing features)
**Success Criteria** (what must be TRUE):
  1. Error messages include actionable suggestions
  2. CI/CD pipeline runs tests on Linux and macOS
  3. Shell wrappers tested on bash 3.2, bash 5+, zsh 5.8+, fish 3.0+
  4. README includes installation, usage, and troubleshooting
**Plans**: 4 plans

Plans:
- [x] 09-01-PLAN.md — Enhanced error messages with fuzzy matching, color output, and help footer
- [x] 09-02-PLAN.md — GitHub Actions CI/CD pipeline (test matrix + release automation)
- [x] 09-03-PLAN.md — End-to-end shell wrapper tests with real git repos
- [x] 09-04-PLAN.md — README rewrite for v2 (installation, commands, configuration, troubleshooting)

</details>

#### Phase 10: UAT Gap Closure
**Goal**: Fix 3 bugs discovered during full UAT: shell wrapper binary resolution, stdout leak in setup executor, and fuzzy matching algorithm
**Depends on**: Phase 9
**Plans**: 2 plans

Plans:
- [ ] 10-01-PLAN.md — Fix stdout leak in setup executor + validate WIP shell wrapper changes
- [ ] 10-02-PLAN.md — Fix fuzzy matching to use segment-aware scoring instead of pure Levenshtein

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Internal Documentation | v1.0 | 3/3 | Complete | 2026-02-07 |
| 2. User-Facing Documentation | v1.0 | 3/3 | Complete | 2026-02-07 |
| 3. Core Go Binary Foundation | v2.0 | 2/2 | Complete | 2026-02-07 |
| 4. Configuration System | v2.0 | 2/2 | Complete | 2026-02-07 |
| 5. Directory-Changing Commands | v2.0 | 3/3 | Complete | 2026-02-07 |
| 6. Shell Integration | v2.0 | 2/2 | Complete | 2026-02-07 |
| 7. npm Distribution | v2.0 | 2/2 | Complete | 2026-02-07 |
| 8. Interactive Installer | v2.0 | 2/2 | Complete | 2026-02-07 |
| 9. Polish & Testing | v2.0 | 4/4 | Complete | 2026-02-08 |
| 10. UAT Gap Closure | v2.0 | 0/2 | In Progress | - |

---
*Roadmap created: 2026-02-07*
*v1.0 complete (phases 1-2), v2.0 phases 3-9 planned*
