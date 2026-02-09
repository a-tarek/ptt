# Roadmap: ptt

## Milestones

- ✅ **v1.0 Documentation** - Phases 1-2 (shipped 2026-02-07)
- ✅ **v2.0 Go Rewrite** - Phases 3-10 (shipped 2026-02-08)
- ✅ **v2.0 Pre-Release Rebrand** - Phases 11-14 (shipped 2026-02-08)
- ✅ **v2.0 Bare Repo + cd Rename** - Phases 15-19 (shipped 2026-02-09)
- **v0.1.2 Release** - Phases 20-21

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
<summary>✅ v2.0 Go Rewrite (Phases 3-10) - SHIPPED 2026-02-08</summary>

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

#### Phase 10: UAT Gap Closure
**Goal**: Fix 3 bugs discovered during full UAT: shell wrapper binary resolution, stdout leak in setup executor, and fuzzy matching algorithm
**Depends on**: Phase 9
**Plans**: 2 plans

Plans:
- [x] 10-01-PLAN.md — Fix stdout leak in setup executor + validate WIP shell wrapper changes
- [x] 10-02-PLAN.md — Fix fuzzy matching to use segment-aware scoring instead of pure Levenshtein

</details>

<details>
<summary>✅ v2.0 Pre-Release Rebrand (Phases 11-14) - SHIPPED 2026-02-08</summary>

### v2.0 Pre-Release Rebrand (Complete)

**Milestone Goal:** Rebrand wt to ptt, restructure commands for better UX, and update distribution before first public release.

#### Phase 11: Go Module + Binary Rename
**Goal**: The Go binary builds and runs as `ptt` with the new module path
**Depends on**: Phase 10
**Requirements**: REN-01, REN-02
**Success Criteria** (what must be TRUE):
  1. User can run `ptt --version` and see version output
  2. User can run `ptt list` and see worktrees (existing commands work under new binary name)
  3. `go build` succeeds with module path `github.com/a-tarek/ptt`
  4. All existing tests pass under the new module path
**Plans**: 1 plan

Plans:
- [x] 11-01-PLAN.md — Update Go module path, all import statements, binary name in command definitions, and build/CI configs

#### Phase 12: Command Restructure + Config Directory
**Goal**: Users interact with the new command names and directory-based config
**Depends on**: Phase 11
**Requirements**: RCMD-01, RCMD-02, RCMD-03, RCMD-04, RCMD-05, CFG-01, CFG-02, CFG-03, CFG-04
**Success Criteria** (what must be TRUE):
  1. User can run `ptt mk <name>` to create a worktree (and `ptt new` still works as alias)
  2. User can run `ptt go <worktree>` to navigate and `ptt go` (no args) to go home
  3. User can run `ptt rm <worktree>` to remove a worktree (and `ptt delete` still works as alias)
  4. User can run `ptt ls` to list worktrees (and `ptt list` still works as alias)
  5. User can run `ptt init` which creates `.pttconfig/default`, and `--config <name>` resolves to `.pttconfig/<name>`
**Plans**: 2 plans

Plans:
- [x] 12-01-PLAN.md — Command restructure: rename commands to mk/go/rm/ls with backward-compatible aliases
- [x] 12-02-PLAN.md — Config directory migration: .pttconfig/ structure with default and named configs

#### Phase 13: Shell Wrappers + npm Distribution
**Goal**: Shell integration and npm packages work under the ptt brand
**Depends on**: Phase 12
**Requirements**: REN-03, REN-04, REN-05, DIST-01, DIST-02, DIST-03
**Success Criteria** (what must be TRUE):
  1. User can source shell wrapper and the `ptt` function is available (bash/zsh/fish)
  2. RC file markers use `>>> ptt >>>` / `<<< ptt <<<` format
  3. User can `npm install @a-tarek/ptt` and get the correct platform binary
  4. Build and publish scripts produce `ptt` binary under `@a-tarek` scope
  5. CI/CD workflow builds and releases the `ptt` binary (not `wt`)
**Plans**: 2 plans

Plans:
- [x] 13-01-PLAN.md — Shell wrapper and installer rebrand (ptt function name, marker blocks, rc file content)
- [x] 13-02-PLAN.md — npm package and CI/CD rebrand (package names, scopes, binary paths, workflow updates)

#### Phase 14: Documentation
**Goal**: README and user-facing docs reflect the ptt brand and new command names
**Depends on**: Phase 13
**Requirements**: DOCS-01
**Success Criteria** (what must be TRUE):
  1. README.md shows `ptt` in all examples, installation instructions, and command reference
  2. README.md documents the new command names (mk, go, rm, ls) with their aliases
  3. README.md installation instructions reference `@a-tarek/ptt`
**Plans**: 1 plan

Plans:
- [x] 14-01-PLAN.md — README rewrite for ptt branding, new command names, and @a-tarek scope

</details>

<details>
<summary>✅ v2.0 Bare Repo + cd Rename (Phases 15-19) - SHIPPED 2026-02-09</summary>

### v2.0 Bare Repo + cd Rename (Complete)

**Milestone Goal:** Add bare repo conversion and nested worktree support, rename go to cd, ship v2.0.

#### Phase 15: Bare Repo Infrastructure
**Goal**: ptt correctly detects bare repo context and resolves worktree paths and config from the right locations
**Depends on**: Phase 14
**Requirements**: BARE-01, BARE-02, BARE-03, BARE-04, BARE-05, BARE-06
**Success Criteria** (what must be TRUE):
  1. Running `ptt mk <name>` inside a ptt bare repo creates the worktree nested inside the container directory (not as a sibling)
  2. Running `ptt init` inside a bare repo worktree creates `.pttconfig/` at the bare repo container root (not inside the worktree)
  3. Running `ptt mk` inside a bare repo resolves config from the container root `.pttconfig/`
  4. Bare repo detection works from any CWD within the bare repo structure (container root, inside a worktree, nested subdirectory)
**Plans**: 2 plans

Plans:
- [x] 15-01-PLAN.md — BareRepoRoot detection, WorktreePath refactor, and ConfigRoot function (TDD)
- [x] 15-02-PLAN.md — Update init, mk, and eject commands to use ConfigRoot for config resolution

#### Phase 16: cd Rename
**Goal**: Users navigate worktrees with `ptt cd` as the primary command, with `go` removed
**Depends on**: Phase 14 (independent of Phase 15)
**Requirements**: CD-01, CD-02, CD-03, CD-04
**Success Criteria** (what must be TRUE):
  1. User can run `ptt cd <worktree>` and the shell changes to that worktree directory
  2. User can run `ptt cd` (no args) and the shell changes to the main worktree
  3. `ptt go` is no longer a recognized command (clean removal, not alias)
  4. Shell wrappers (bash/zsh/fish) handle the `cd` subcommand for directory changes
**Plans**: 1 plan

Plans:
- [x] 16-01-PLAN.md -- Rename go command to cd across Go source, shell wrappers, tests, and README

#### Phase 17: mk-bare-repo Command
**Goal**: Users can convert any normal clone into a bare repo layout with a single command
**Depends on**: Phase 15
**Requirements**: MKBR-01, MKBR-02, MKBR-03, MKBR-04, MKBR-05, MKBR-06, MKBR-07, MKBR-08
**Success Criteria** (what must be TRUE):
  1. User can run `ptt mk-bare-repo` in a normal clone and a `<repo>-bare/` sibling directory is created with the correct bare layout (`.bare/` + `.git` pointer)
  2. User can `cd` into the new bare repo and run `git fetch` successfully (refspec configured correctly)
  3. The new bare repo contains an initial worktree checked out to the default branch (main or master)
  4. If the source repo has `.pttconfig/`, it is copied to the new bare repo container root
  5. Running `ptt mk-bare-repo` in an already-converted bare repo or a repo without remotes produces a clear error
**Plans**: 2 plans

Plans:
- [x] 17-01-PLAN.md -- mk-bare-repo command implementation and integration tests
- [x] 17-02-PLAN.md -- README documentation for mk-bare-repo

#### Phase 18: Adopt + Smart Init
**Goal**: `ptt init` becomes the single smart command for initializing any repo into ptt-managed bare layout
**Depends on**: Phase 15
**Requirements**: ADOPT-01, ADOPT-02, ADOPT-03, INIT-01
**Success Criteria** (what must be TRUE):
  1. User can run `ptt init` inside a raw `git clone --bare` repo and it restructures into ptt layout (`.bare/` wrapper, `.git` pointer, default branch worktree)
  2. After adoption, `git fetch` works correctly in the restructured repo (refspec and reflog configured)
  3. User can run `ptt init` in a normal clone and it restructures in-place to ptt bare layout with nested worktrees
  4. `mk-bare-repo` command removed -- `ptt init` replaces it entirely
**Plans**: 3 plans

Plans:
- [x] 18-01-PLAN.md -- Remove mk-bare-repo and create detection/validation/plan display infrastructure
- [x] 18-02-PLAN.md -- Implement restructuring operations (normal clone, raw bare adoption, repair)
- [x] 18-03-PLAN.md -- Rewrite init command with context routing and integration tests

#### Phase 19: Polish
**Goal**: Bare repo metadata is hidden from user-facing output
**Depends on**: Phase 17
**Requirements**: POL-01
**Success Criteria** (what must be TRUE):
  1. Running `ptt ls` inside a bare repo shows only real worktrees, not the `.bare` metadata entry
  2. The bare metadata entry is filtered regardless of how git reports it (path or label variations)
**Plans**: 1 plan

Plans:
- [x] 19-01-PLAN.md -- Filter bare repo metadata entries from ptt ls output

</details>

### v0.1.2 Release (Phases 20-21)

**Milestone Goal:** Cheatsheet-style README and first published release (GitHub + npm).

#### Phase 20: README Rewrite
**Goal**: README is a one-screen cheatsheet with SVG banner and concise command reference
**Depends on**: Phase 19
**Requirements**: README-01, README-02, README-03, README-04, README-05, README-06
**Success Criteria** (what must be TRUE):
  1. User sees SVG banner at top of README (from assets/banner.svg)
  2. User sees one-line project description immediately after banner
  3. User sees npm install command (`npm install @a-tarek/ptt`) in installation section
  4. User sees core commands (mk, cd, rm, ls) with one-liner examples in command reference
  5. User sees `ptt init` bare repo workflow documented with example
  6. README fits approximately one screen — no verbose explanations or separate docs files
**Plans**: TBD

Plans:
- [ ] 20-01-PLAN.md — README rewrite: SVG banner, one-line description, npm install, core commands with examples, init workflow

#### Phase 21: v0.1.2 Release
**Goal**: v0.1.2 is published to GitHub and npm with downloadable binaries
**Depends on**: Phase 20
**Requirements**: REL-01, REL-02, REL-03
**Success Criteria** (what must be TRUE):
  1. Git tag `v0.1.2` exists and is pushed to GitHub
  2. GitHub release exists at https://github.com/a-tarek/ptt/releases/tag/v0.1.2 with platform binaries attached
  3. Package `@a-tarek/ptt` is published to npm registry with version 0.1.2
  4. Running `npm install @a-tarek/ptt` downloads version 0.1.2
**Plans**: TBD

Plans:
- [ ] 21-01-PLAN.md — Create git tag v0.1.2, trigger goreleaser workflow for GitHub release
- [ ] 21-02-PLAN.md — Publish @a-tarek/ptt@0.1.2 to npm registry

## Progress

**Execution Order:**
Phase 20 must complete before Phase 21 (README must be ready before release).

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
| 10. UAT Gap Closure | v2.0 | 2/2 | Complete | 2026-02-08 |
| 11. Go Module + Binary Rename | Rebrand | 1/1 | Complete | 2026-02-08 |
| 12. Command Restructure + Config Directory | Rebrand | 2/2 | Complete | 2026-02-08 |
| 13. Shell Wrappers + npm Distribution | Rebrand | 2/2 | Complete | 2026-02-08 |
| 14. Documentation | Rebrand | 1/1 | Complete | 2026-02-08 |
| 15. Bare Repo Infrastructure | Bare Repo + cd | 2/2 | Complete | 2026-02-09 |
| 16. cd Rename | Bare Repo + cd | 1/1 | Complete | 2026-02-09 |
| 17. mk-bare-repo Command | Bare Repo + cd | 2/2 | Complete | 2026-02-09 |
| 18. Adopt + Smart Init | Bare Repo + cd | 3/3 | Complete | 2026-02-09 |
| 19. Polish | Bare Repo + cd | 1/1 | Complete | 2026-02-09 |
| 20. README Rewrite | v0.1.2 | 0/1 | Pending | - |
| 21. v0.1.2 Release | v0.1.2 | 0/2 | Pending | - |

---
*Roadmap created: 2026-02-07*
*v1.0 complete (phases 1-2), v2.0 complete (phases 3-10), rebrand complete (phases 11-14), bare repo + cd complete (phases 15-19)*
*v0.1.2 release milestone added: 2026-02-09*
