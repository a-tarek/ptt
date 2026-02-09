# Project Research Summary

**Project:** ptt -- Bare Repo Support + `cd` Rename Milestone
**Domain:** Git worktree manager CLI tool -- bare repo conversion and nested worktree workflows
**Researched:** 2026-02-09
**Confidence:** HIGH

## Executive Summary

This milestone adds bare repo support to ptt, enabling the community-standard pattern where a container directory holds a hidden `.bare/` git database alongside sibling worktree directories. The ecosystem has converged on this layout: a `project-bare/` container with `.bare/` (the git database), a `.git` file (pointer to `.bare/`), and worktrees as sibling directories (`main/`, `feature-x/`). Eight competing tools were surveyed, with git-prole's `convert` command being the closest analog to ptt's planned `mk-bare-repo`. The codebase already has partial bare repo awareness -- `IsBareRepository()`, `GetRepoRoot()`, `GetHomePath()`, and `WorktreePath()` in `internal/git/repo.go` already detect bare repos and branch behavior. The primary gap is a new `BareRepoRoot()` function that reliably identifies the ptt container directory from any nested worktree.

The recommended approach centers on a clone-from-remote strategy: `ptt mk-bare-repo` creates a new `project-bare/` sibling directory by running `git clone --bare <remote-url>` rather than restructuring `.git/` in place. This is safer (original repo untouched), simpler (no git internals manipulation), and idempotent (user can retry after failure). The key architectural insight is that bare repo awareness stays exclusively in `internal/git/repo.go` -- the config package remains unaware of repo type, receiving only a root path from a new `ConfigRoot()` convenience function. The `cd` rename (from `go` to `cd`) is fully independent and can be implemented in parallel with any bare repo phase.

The main risks are well-understood: (1) `WorktreePath()` has fragile bare detection using `strings.HasSuffix(repoRoot, ".git")` that does not match ptt's naming convention -- fix by replacing with `BareRepoRoot()` call; (2) `git clone --bare` silently breaks `git fetch` by not setting `remote.origin.fetch` -- fix by explicitly configuring the refspec during setup; (3) `.pttconfig/` must live at the container root for bare repos, not inside any worktree -- fix by updating all config resolution callers to use `ConfigRoot()`. All three have clear, surgical fixes documented in the architecture research.

## Key Findings

### Recommended Stack

No new libraries or dependencies are needed. All changes use existing packages (`os/exec` for git commands, `os` for filesystem operations) and patterns already present in the codebase.

**New functions to add to `internal/git/repo.go`:**
- `BareRepoRoot()` -- detect ptt bare repo structure via `git rev-parse --git-common-dir` then check parent directory has a `.git` file (not directory); this is the foundation everything else depends on
- `ConfigRoot()` -- convenience wrapper returning `BareRepoRoot()` if bare, else `GetHomePath()`; centralizes the "where does .pttconfig/ live?" decision for all callers
- `GetRemoteURL()` -- shell out to `git remote get-url origin`; needed by `mk-bare-repo`
- `GetDefaultBranch()` -- detect main/master as default branch; needed by `mk-bare-repo`

**New command file:**
- `cmd/mk_bare_repo.go` -- cobra command for bare repo conversion

### Expected Features

**Must have (table stakes):**
- Bare repo detection -- ptt must know it is in a bare repo to change worktree placement
- Nested worktree creation -- in bare repos, worktrees go inside container dir, not as siblings
- Correct fetch config -- `git clone --bare` breaks `git fetch` without explicit `remote.origin.fetch` refspec
- `.git` pointer file creation -- container needs `gitdir: ./.bare` for git commands to work
- Navigate to home worktree -- `ptt cd` (no args) must find the right worktree, not the bare dir (already works via `GetHomePath()`)
- `ptt ls` in bare repos -- already works via porcelain parsing
- Config resolution in bare repos -- `.pttconfig/` at container root, not inside worktrees
- Initial worktree for default branch -- after bare clone, create `main/` checkout automatically
- Enable `core.logallrefupdates` -- bare repos disable reflog by default, surprising users

**Should have (differentiators):**
- `mk-bare-repo` conversion command -- convert existing clone to bare layout by cloning from remote into new sibling directory; only git-prole offers comparable functionality
- `.pttconfig/` at container level -- natural shared config location for all nested worktrees
- Branch name slash-to-dash conversion -- `feature/auth` becomes `feature-auth/` directory name
- `ls` filters bare entry -- hide `.bare` metadata entry from worktree listings
- `cd` rename -- rename primary navigation command from `go` to `cd` with `go` as backward-compatible alias

**Defer (v2+):**
- Dirty working tree handling during conversion (stash/restore) -- clone-from-remote is sufficient for v1
- Existing worktree migration during conversion -- error with guidance instead
- Worktree cleanup/prune command -- useful for long-lived bare repos but not launch-critical
- In-place conversion of `.git/` directory -- always create new sibling, never mutate original

**Anti-features (deliberately not building):**
- `ptt clone` command -- duplicates `git clone --bare` with minimal added value; `mk-bare-repo` is the harder, more valuable workflow
- Wrapping `git checkout`/`git branch` -- keep explicit `ptt mk`/`ptt cd`/`ptt rm`
- Global worktree registry -- ptt operates within a single repo context
- Tmux/editor integration -- stay editor-agnostic
- Interactive TUI for worktree selection -- stay CLI-first
- Destroy command (delete remote branch) -- too destructive for a convenience tool
- Force-convert in-place -- always create new sibling directory, leave original untouched
- Dotfiles management mode -- orthogonal use case

### Architecture Approach

The integration is surgical. Bare repo awareness lives exclusively in `internal/git/repo.go` via one new detection function (`BareRepoRoot()`) and one convenience function (`ConfigRoot()`). The config package (`internal/config/`) remains completely unaware of repo type -- callers simply pass the correct root path. The `mk-bare-repo` command is a new cobra command that orchestrates a sequence of standard git commands. No new internal packages are needed, and 15+ existing components remain completely unchanged.

**Major components and their responsibilities:**

1. **`BareRepoRoot()` in `internal/git/repo.go`** -- the foundation; uses `git rev-parse --git-common-dir` to find `.bare/`, then checks parent has a `.git` file (not directory) as the ptt bare repo signal
2. **`ConfigRoot()` in `internal/git/repo.go`** -- returns bare repo root or home worktree path; all config resolution callers use this instead of `GetHomePath()`
3. **Fixed `WorktreePath()` in `internal/git/repo.go`** -- replace fragile `strings.HasSuffix` heuristic with `BareRepoRoot()` call; in bare mode, `filepath.Join(bareRoot, name)` instead of sibling mode
4. **`cmd/mk_bare_repo.go`** -- new command: validate (not bare, has remote) -> `git clone --bare` -> write `.git` pointer -> config fetch refspec -> enable reflog -> fetch -> `git worktree add` -> copy `.pttconfig/`
5. **Config resolution callers** (`cmd/new.go`, `cmd/eject.go`, `cmd/init_cmd.go`) -- switch from `GetHomePath()` to `ConfigRoot()` for config path only; source root for copy/symlink stays as current worktree
6. **`cmd/goto.go`** -- rename `Use` from `go` to `cd`, add `go` as backward-compatible alias
7. **Shell wrapper templates** (bash/zsh/fish) -- add `cd` to case list alongside existing `go|goto|home`

**Key architectural decision: config package stays repo-type-agnostic.** `ResolveConfigPath(repoRoot, name)` is correct as-is. The bare repo awareness belongs in callers and in `internal/git/`, not in the config package. This keeps the abstraction boundary clean.

### Critical Pitfalls

1. **`WorktreePath()` bare detection is fragile** -- current code uses `strings.HasSuffix(repoRoot, ".git")` which does not match ptt's bare repo naming (`project-bare/`). Must replace with `BareRepoRoot()` call using the authoritative `git rev-parse --git-common-dir` + `.git`-file-not-directory check. This is the first thing to fix.

2. **`git clone --bare` silently breaks fetch** -- bare clones do not set `remote.origin.fetch`, so `git fetch` appears to succeed but fetches nothing. The fix is a single config set during `mk-bare-repo` setup: `remote.origin.fetch = +refs/heads/*:refs/remotes/origin/*`. Well-documented across multiple community sources.

3. **Config resolution root mismatch in bare repos** -- `.pttconfig/` lives at the container root (`project-bare/.pttconfig/`), but current callers pass `GetHomePath()` (which returns the main worktree path, e.g., `project-bare/main/`). Fix: all callers use new `ConfigRoot()`. The config package itself does not change.

4. **Nested path calculation depends on correct root** -- `WorktreePath()` does `filepath.Dir(repoRoot)` then `filepath.Join(parentDir, name)`. Whether this works depends on what `repoRoot` actually is. Using `BareRepoRoot()` explicitly returns the container directory, making `filepath.Join(bareRoot, name)` always correct.

5. **Bare repos disable reflog by default** -- `core.logallrefupdates` defaults to false in bare repos, which surprises users who rely on reflog for recovery. Fix: set `core.logallrefupdates=true` during `mk-bare-repo` setup.

## Implications for Roadmap

Based on dependency analysis from the architecture research, the suggested phase structure follows the build order: foundation functions first (everything depends on `BareRepoRoot()`), then independent `cd` rename, then config resolution (needed before the new command), then the command itself, then polish.

### Phase 1: Bare Repo Detection Foundation

**Rationale:** Everything depends on `BareRepoRoot()`. No bare repo feature works without correct detection. This phase has zero user-visible changes, making it safe to ship incrementally. The detection algorithm is well-defined: `git rev-parse --git-common-dir` returns `.bare/` path, `filepath.Dir()` gives container root, check parent has `.git` file (not directory).
**Delivers:** `BareRepoRoot()`, `ConfigRoot()`, fixed `WorktreePath()` bare detection and path calculation
**Addresses:** Bare repo detection (table stakes), nested worktree path resolution (table stakes)
**Avoids:** Fragile `WorktreePath()` heuristic (pitfall 1), path calculation off-by-one (pitfall 4)
**Components:** `internal/git/repo.go` only -- pure additions plus one function fix

### Phase 2: `cd` Rename

**Rationale:** Fully independent of all bare repo work. Zero dependencies on Phase 1. Can run in parallel. Smallest scope of any phase -- rename cobra command `Use` field, add alias, update shell wrappers and help text. Ships a user-visible improvement immediately.
**Delivers:** `ptt cd` as primary navigation command, `ptt go` as backward-compatible alias
**Addresses:** `cd` rename feature
**Avoids:** No pitfalls. Shell builtin `cd` is not in conflict because `cd` is a subcommand of `ptt`, not a standalone command. The wrapper calls `cd "$result"` internally, which invokes the builtin.
**Components:** `cmd/goto.go` (rename Use, add alias), `cmd/root.go` (help text), `internal/shell/templates/wrapper.{bash,zsh,fish}` (add `cd` to case list)

### Phase 3: Config Resolution for Bare Repos

**Rationale:** Depends on Phase 1 (`ConfigRoot()` calls `BareRepoRoot()`). Must be done before Phase 4 because `mk-bare-repo` copies `.pttconfig/` to the container root, and existing commands must be able to find it there.
**Delivers:** `ptt mk`, `ptt eject`, `ptt init` correctly resolve `.pttconfig/` from bare repo container root
**Addresses:** Config file resolution in bare repos (table stakes), `.pttconfig/` at container level (differentiator)
**Avoids:** Config resolution root mismatch (pitfall 3)
**Components:** `cmd/new.go`, `cmd/eject.go`, `cmd/init_cmd.go` -- callers only; `internal/config/resolve.go` is unchanged

### Phase 4: `mk-bare-repo` Command

**Rationale:** The primary user-facing deliverable. Depends on Phase 1 (detection) and Phase 3 (config resolution). This is the highest-complexity phase but the conversion flow is a linear sequence of standard git commands.
**Delivers:** `ptt mk-bare-repo` command that creates a bare repo layout from an existing clone
**Addresses:** `mk-bare-repo` conversion (differentiator), fetch config (table stakes), `.git` pointer (table stakes), initial worktree (table stakes), reflog enable (table stakes)
**Avoids:** Silent fetch breakage (pitfall 2), reflog disabled (pitfall 5), in-place mutation anti-pattern (always creates new sibling directory)
**Components:** `cmd/mk_bare_repo.go` (new), `internal/git/repo.go` (new helpers: `GetRemoteURL()`, `GetDefaultBranch()`)

### Phase 5: Polish and Integration Testing

**Rationale:** Depends on all above. Covers filtering, edge case handling, and end-to-end testing of the complete bare repo workflow.
**Delivers:** Clean `ptt ls` output (bare entry filtered), branch slash-to-dash handling, comprehensive integration tests
**Addresses:** `ls` bare entry filtering, edge cases (running from container root, deleted main worktree, branch names with slashes, submodule warnings)
**Components:** `cmd/list.go`, integration tests, documentation updates

### Phase Ordering Rationale

- **Phase 1 first** because `BareRepoRoot()` is the dependency for Phases 3, 4, and 5
- **Phase 2 is independent** and can run in parallel with Phase 1; placed second for logical ordering but has zero dependencies on bare repo work
- **Phase 3 before Phase 4** because `mk-bare-repo` copies `.pttconfig/` to the container root, and `mk`/`eject`/`init` must already know to look there
- **Phase 4 is the big deliverable** and benefits from all foundation work being complete
- **Phase 5 last** because polish and testing should not introduce new logic; it validates everything built in Phases 1-4

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 1:** Needs integration testing to verify `git rev-parse --git-common-dir` behavior from all CWD locations (inside worktree, container root, `.bare/` itself). Algorithm is sound but edge cases need real temp-dir tests.
- **Phase 4:** The `mk-bare-repo` command has edge cases: repos without remotes (error), repos with existing linked worktrees (error with guidance), submodules (warn). V1 approach is to error on complex cases rather than handle them, but error paths need verification.

Phases with standard patterns (skip research-phase):
- **Phase 2:** Trivial cobra command rename + shell wrapper update. No research needed.
- **Phase 3:** Mechanical caller updates -- pass different root to unchanged function. No research needed.
- **Phase 5:** Standard testing and documentation. No research needed.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | No new dependencies; all changes use existing packages and patterns from the codebase |
| Features | HIGH | Comprehensive ecosystem survey of 8+ tools; clear consensus on table stakes vs anti-features; canonical bare repo pattern is well-documented |
| Architecture | HIGH | Based on direct source code analysis of `internal/git/repo.go`, `cmd/new.go`, `cmd/goto.go`, `cmd/eject.go`, `cmd/init_cmd.go`, shell wrappers |
| Pitfalls | HIGH | All pitfalls identified from codebase analysis + verified against git official documentation + corroborated by multiple community sources |

**Overall confidence:** HIGH

### Gaps to Address

- **`git rev-parse --git-common-dir` from container root:** Detection algorithm is sound, but exact behavior when CWD is the container root (with `.git` file but not inside any worktree) needs integration testing during Phase 1. Low risk -- `.git` file makes git recognize the directory.
- **`git worktree add` relative path resolution from nested worktree:** When running from inside `project-bare/main/`, does git correctly resolve relative paths for new worktrees? Needs verification during Phase 1 tests. Workaround: always use absolute paths in the `git worktree add` command.
- **Submodule behavior during bare clone:** `git clone --bare` may not correctly handle submodules. V1 approach: warn and skip. Needs testing during Phase 4.
- **Repos without remotes:** `mk-bare-repo` requires a remote URL to clone from. Purely local repos cannot be converted. This is an acceptable constraint (bare repos are most useful for remote-tracked repos) but the error message must be clear.

## Sources

### Primary (HIGH confidence)
- [Git rev-parse documentation](https://git-scm.com/docs/git-rev-parse) -- `--git-common-dir`, `--is-bare-repository`, `--git-dir` behavior
- [Git worktree documentation](https://git-scm.com/docs/git-worktree) -- worktree list, bare repo display, linked worktree mechanics
- [Git clone documentation](https://git-scm.com/docs/git-clone) -- `--bare` flag behavior and limitations
- [Git config documentation](https://git-scm.com/docs/git-config) -- worktree config scope, bare repo config
- Direct source code analysis: `internal/git/repo.go`, `cmd/new.go`, `cmd/goto.go`, `cmd/eject.go`, `cmd/init_cmd.go`, `internal/config/resolve.go`, shell wrapper templates

### Secondary (MEDIUM confidence)
- [Morgan Cugerone: How to use git worktree in a clean way](https://morgan.cugerone.com/blog/how-to-use-git-worktree-and-in-a-clean-way/) -- canonical `.bare/` + `.git` pointer pattern
- [Morgan Cugerone: Workarounds for bare repo fetch](https://morgan.cugerone.com/blog/workarounds-to-git-worktree-using-bare-repository-and-cannot-fetch-remote-branches/) -- `remote.origin.fetch` fix
- [Andreas Schneider: git-worktree and bare repo](https://blog.cryptomilk.org/2023/02/10/sliced-bread-git-worktree-and-bare-repo/) -- nested worktree pattern
- [git-prole announcement](https://becca.ooo/blog/announcing-git-prole/) -- `convert` command design, bare repo layout
- [matklad: How I Use Git Worktrees](https://matklad.github.io/2024/07/25/git-worktrees.html) -- fixed worktree set pattern
- [Nick Nisi: How I use git worktrees](https://nicknisi.com/posts/git-worktrees/) -- real-world bare repo usage
- Ecosystem survey: git-wt, git-worktree-wrapper, wtm, gwq, wtp, worktrunk, gtr, git-prole

### Tertiary (LOW confidence, needs validation)
- Exact `git rev-parse --git-common-dir` output when CWD is container root with `.git` file -- needs integration test
- Submodule behavior during bare clone -- needs testing during implementation
- Whether `git worktree add` from inside a nested worktree resolves relative paths to container root -- needs integration test

---
*Research completed: 2026-02-09*
*Ready for roadmap: yes*
