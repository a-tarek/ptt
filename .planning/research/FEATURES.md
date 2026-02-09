# Feature Landscape: Bare Repo Conversion & Nested Worktree Workflows

**Domain:** Git worktree manager CLI tool -- bare repo support milestone
**Researched:** 2026-02-09
**Applies to:** ptt v3.0 (bare repo + nested worktree features)

---

## Ecosystem Survey: How Bare Repo Worktree Workflows Work in the Wild

### The Canonical Bare Repo + Worktree Pattern

The community has converged on a standard layout for bare repo worktree workflows. This pattern appears across virtually all tools and blog posts surveyed:

```
my-project/              # container directory
  .bare/                 # bare git repo (all metadata)
  .git                   # file (not directory) containing: gitdir: ./.bare
  main/                  # worktree: default branch
  feature-auth/          # worktree: feature branch
  hotfix-security/       # worktree: hotfix branch
```

Key characteristics:
1. **`.bare/` hides git internals** -- the bare clone lives in a hidden directory
2. **`.git` is a file, not a directory** -- contains `gitdir: ./.bare` pointer
3. **Worktrees are siblings** -- nested inside the container directory, not outside it
4. **Each worktree is a full checkout** -- has its own `.git` file pointing back to `.bare/worktrees/<name>`

This is fundamentally different from ptt's current "sibling mode" where worktrees are created as siblings to the repo (`repo-staging/` next to `repo/`). In bare repo mode, worktrees live **inside** the container directory.

### Existing Tools Surveyed

| Tool | Language | Bare Focus | Key Differentiator |
|------|----------|------------|-------------------|
| [git-wt (gabri.me)](https://gabri.me/blog/git-wt) | Shell | Yes | `clone`, `migrate`, `switch` with fzf, `destroy` (remote cleanup) |
| [git-worktree-wrapper](https://github.com/lu0/git-worktree-wrapper) | Shell | Yes | Wraps `git checkout`/`git branch` to auto-create worktrees |
| [worktree-manager (wtm)](https://github.com/jarredkenny/worktree-manager) | TypeScript/Bun | Yes | `post_create` hooks, `cleanup` for merged branches |
| [gwq](https://github.com/d-kuro/gwq) | Go | No | Global worktree management, tmux integration, status dashboard |
| [wtp](https://github.com/satococoa/wtp) | Go | No | `.wtp.yml` config with copy/symlink/command hooks |
| [worktrunk](https://github.com/max-sixty/worktrunk) | Rust | No | AI agent parallel workflows, LLM commit messages |
| [git-worktree-runner (gtr)](https://github.com/coderabbitai/git-worktree-runner) | Bash | No | Editor/AI tool integration, `.gtrconfig` |
| [git-prole](https://becca.ooo/blog/announcing-git-prole/) | Rust | Yes | `convert` existing repos, auto-copies untracked files |

### What Users Expect (Behavioral Patterns)

From blog posts and tool READMEs, the standard user expectations are:

1. **One-command bare clone setup** -- nobody wants to run 5+ manual commands
2. **`git fetch` works out of the box** -- bare clones break this without `remote.origin.fetch` config
3. **Worktrees inside the container** -- not scattered as siblings
4. **Navigate by branch name, not path** -- users think in branches, not directories
5. **Config files shared or copied** -- `.env`, `node_modules`, `.venv` handled automatically
6. **Easy cleanup** -- remove worktree + optionally delete branch + optionally delete remote branch

---

## Table Stakes

Features users absolutely expect for bare repo workflow support. Missing any of these makes the feature feel broken or incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **Bare repo detection** | ptt must know it's in a bare repo to change worktree placement | Low | Already partially implemented (`IsBareRepository()`, `WorktreePath()`) |
| **Nested worktree creation** | In bare repos, worktrees go inside container dir, not as siblings | Low | `WorktreePath()` already has this logic (`parentDir/name` for bare) |
| **Correct fetch config** | `git clone --bare` does not set `remote.origin.fetch`, breaking `git fetch` | Low | Must set `+refs/heads/*:refs/remotes/origin/*` during setup |
| **`.git` pointer file creation** | Container dir needs `gitdir: ./.bare` file for git commands to work | Low | Single file write: `echo "gitdir: ./.bare" > .git` |
| **Navigate to home worktree in bare repo** | `ptt cd` (no args) must find the right worktree, not the bare dir | Medium | Already implemented: `GetHomePath()` finds non-bare worktree matching HEAD |
| **`ptt ls` works in bare repos** | Listing worktrees must work regardless of repo type | Low | Already works -- `ListWorktrees()` parses porcelain output correctly |
| **Config file resolution in bare repos** | `.pttconfig/` must be findable from any nested worktree | Medium | Currently resolves from `homePath` -- in bare repos, must resolve from container root or main worktree |
| **Initial worktree for default branch** | After bare conversion/clone, user needs at least one working checkout | Low | Create `main/` (or whatever HEAD points to) worktree automatically |
| **Enable `core.logallrefupdates`** | Bare repos disable reflog by default, which surprises users | Low | One config set during setup |

## Differentiators

Features that set ptt apart from alternatives. These are valuable but not strictly required for bare repo support to function.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **`ptt mk-bare` conversion command** | Convert existing normal clone to bare layout without re-cloning | High | git-prole's `convert` is its killer feature; most tools only do `clone --bare` |
| **Preserve uncommitted work during conversion** | Stash/restore changes when converting to bare layout | Medium | Users will try to convert dirty repos -- must handle gracefully |
| **`.pttconfig/` at container level** | Config in container root applies to ALL nested worktrees | Low | Natural location for bare repos -- container root is the shared context |
| **Branch name as worktree name** | `ptt mk feature/auth` creates `feature-auth/` worktree (slash to dash) | Low | wtp and git-wt both do this; cleaner than raw git |
| **Post-create hook/config for bare repos** | Auto-run `npm install` or similar after worktree creation | Medium | wtm and wtp both offer this; ptt already has `run` action type |
| **Worktree cleanup/prune** | Remove worktrees whose branches have been merged | Medium | wtm has `cleanup` command; useful for long-lived bare repos |
| **`ptt cd` no-arg behavior in bare repos** | Navigate to the default branch worktree (not the bare dir) | Low | Already implemented in `GetHomePath()` |
| **Fuzzy worktree switching** | `ptt cd` with partial name matching | Low | Already implemented via `ResolveWorktree()` with fuzzy scoring |

## Anti-Features

Features to deliberately NOT build. These are common in the ecosystem but are wrong choices for ptt.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **Wrap `git checkout`/`git branch`** | git-worktree-wrapper's approach of intercepting standard git commands creates confusion and unexpected behavior | Keep explicit commands (`ptt mk`, `ptt cd`, `ptt rm`) -- users should know they're managing worktrees |
| **`ptt clone` command** | Duplicates `git clone --bare` with minimal added value; users already know how to clone | Provide `ptt mk-bare` to convert existing repos; that is the harder workflow to get right |
| **Global worktree registry** | gwq tracks worktrees across all repos globally; adds persistence, state management, and complexity | Keep per-repo scope -- ptt operates within a single repo context |
| **Tmux/editor integration** | gwq and gtr integrate with tmux/VS Code; adds coupling to specific tools | Stay editor-agnostic -- ptt outputs paths, users wire up their own editor workflows |
| **LLM/AI commit messages** | worktrunk generates commit messages from diffs; orthogonal to worktree management | Out of scope -- let users pick their own commit message tooling |
| **Interactive TUI for worktree selection** | fzf-based pickers are nice but add dependencies and complexity | Keep CLI-first; users who want fzf can pipe `ptt ls` through it themselves |
| **Destroy command (delete remote branch)** | git-wt's `destroy` deletes worktree + local branch + remote branch; too destructive for a convenience tool | `ptt rm` removes worktree only; branch deletion is a deliberate git operation users should do explicitly |
| **Automatic dependency installation** | gtr and wtp auto-run `npm install` after worktree creation | ptt already supports `run` actions in `.pttconfig/` -- users configure what they need rather than ptt guessing |
| **Force-convert in-place** | Converting a repo in-place (mutating `.git/` to bare) risks data loss | Always create a new container directory (`repo-bare/`), leaving original untouched until user verifies |
| **Dotfiles management mode** | Some bare repo tools support the `$HOME` as worktree pattern for dotfiles | Orthogonal use case -- ptt is for project development, not dotfiles management |
| **PR/CI integration** | worktrunk shows CI status and PR links; scope creep | Let `gh` CLI handle PR workflows; ptt manages worktrees |

---

## Detailed Feature Analysis

### Feature 1: `ptt mk-bare` -- Convert Existing Repo to Bare Layout

**What it does:** Takes a normal `git clone` and restructures it into the bare repo + nested worktree layout.

**Why it matters:** This is the hardest part of bare repo adoption. Users have existing repos they want to convert without re-cloning (which loses local branches, stashes, reflog). Only git-prole offers this, and it is cited as its primary value.

**Expected behavior:**
```
$ cd ~/code/my-project      # normal clone, currently on main
$ ptt mk-bare

Creating bare repo layout...
  create: my-project-bare/
  move:   .git -> my-project-bare/.bare/
  create: my-project-bare/.git (pointer)
  config: remote.origin.fetch
  config: core.logallrefupdates
  add:    my-project-bare/main/ (worktree)
  copy:   working tree -> main/

Ready: ~/code/my-project-bare/main/
```

**Critical decisions:**
- Create sibling directory (`my-project-bare/`) vs in-place conversion
- Recommendation: **sibling directory** -- safer, user can verify before deleting original
- Handle dirty working tree: stash first, restore in new worktree
- Handle existing worktrees: must migrate them or error with guidance
- Preserve `.pttconfig/`: copy to container root

**Complexity:** HIGH -- many edge cases (dirty tree, existing worktrees, submodules, hooks)

### Feature 2: `ptt mk` in Bare Repos -- Nested Worktree Creation

**What it does:** When inside a bare repo layout, `ptt mk <name>` creates a worktree as a subdirectory of the container rather than a sibling.

**Current state:** Already partially implemented. `WorktreePath()` in `repo.go` detects bare repos and returns `parentDir/name` instead of `parentDir/repoName-name`. However, "parentDir" resolution in bare repos needs verification.

**Expected behavior:**
```
$ pwd
~/code/my-project-bare/main

$ ptt mk feature-auth

  create: feature-auth
  copy:   .env.local
  symlink: node_modules
  cd:     feature-auth

Ready: ~/code/my-project-bare/feature-auth/
Branch: feature-auth
```

**Key subtlety:** The container directory is the parent of both `.bare/` and the worktrees. When running from inside a worktree (e.g., `main/`), ptt must resolve "up" to the container directory, then create the new worktree there.

**Path resolution chain:**
1. `git rev-parse --show-toplevel` returns current worktree root (e.g., `~/code/my-project-bare/main`)
2. Need to find container root: parent of that worktree
3. In bare repo layout: container = `dirname(worktree_root)` = `~/code/my-project-bare/`
4. New worktree goes at `container/name` = `~/code/my-project-bare/feature-auth/`

### Feature 3: Config File Resolution in Bare Repos

**Current behavior:** `ResolveConfigPath()` looks for `.pttconfig/default` relative to `homePath` (the main worktree path). In a normal repo, this works because `.pttconfig/` lives in the repo root. In a bare repo, the question becomes: where does `.pttconfig/` live?

**Two options:**

| Location | Pros | Cons |
|----------|------|------|
| Container root (`my-project-bare/.pttconfig/`) | Shared across all worktrees naturally; one config for the whole project | Not inside any git-tracked directory |
| Main worktree (`my-project-bare/main/.pttconfig/`) | Git-tracked; travels with the repo | Other worktrees must look into `main/` to find config; breaks if main worktree is deleted |

**Recommendation:** Container root. Rationale:
- `.pttconfig/` is already gitignored (it contains local paths like `.env`)
- The container directory IS the project root conceptually
- Config applies to worktree creation, which is a container-level operation
- Consistent: always look at the level where you'd run `ptt mk`

**Resolution chain for bare repos:**
1. Detect bare repo
2. Find container root (parent of `.bare/`)
3. Look for `.pttconfig/default` at container root

**Fallback:** If not found at container root, check main worktree root. This handles repos where `.pttconfig/` was committed and is part of the tracked tree.

### Feature 4: `ptt cd` Navigation in Bare Repos

**Current behavior:** `ptt cd` (no args) calls `GetHomePath()` which finds the first non-bare worktree matching the bare repo's HEAD branch. This already works correctly.

**`ptt cd <name>` behavior:** `ResolveWorktree()` uses suffix matching on directory basenames. In bare repos, worktree basenames ARE the names (e.g., `feature-auth`), not `repo-feature-auth`. This means the name IS the full basename -- no suffix stripping needed.

**Current resolve logic works for both modes:**
- Normal repo: `basename == name` OR `basename ends with -name` (suffix match)
- Bare repo: `basename == name` (direct match -- worktrees are named by branch)

**No changes needed** for basic navigation. The existing resolve logic handles both cases.

---

## Feature Dependencies

```
mk-bare (conversion)
  |
  +-- bare repo detection (already exists)
  +-- .git pointer file creation
  +-- fetch config setup
  +-- logallrefupdates config
  +-- initial worktree creation
  |
  +---> mk in bare repos (nested worktree creation)
  |       +-- path resolution (already exists, needs verification)
  |       +-- config resolution in bare repos
  |       +-- .pttconfig/ at container root
  |
  +---> cd in bare repos (navigation)
  |       +-- GetHomePath() (already works)
  |       +-- ResolveWorktree() (already works)
  |
  +---> ls in bare repos (listing)
          +-- already works via porcelain parsing
```

## Priority Order for Implementation

1. **Bare repo detection + nested path resolution** (verify existing code handles all cases)
2. **Config resolution in bare repos** (update `ResolveConfigPath` for container root)
3. **`ptt mk-bare`** (the conversion command -- the big new feature)
4. **`ptt cd` rename** (from `go` to `cd` -- separate from bare repo work)
5. **Post-creation hooks in bare repos** (verify `run` actions work correctly)

## Edge Cases

| Edge Case | How to Handle |
|-----------|---------------|
| Convert repo with uncommitted changes | Stash, convert, restore stash in new main worktree |
| Convert repo that already has linked worktrees | Error: "repo has existing worktrees -- remove them first or use --force" |
| `ptt mk-bare` when already in a bare repo | Error: "already in a bare repo layout" |
| Container directory name collision | Error: "directory already exists: my-project-bare/" |
| Running `ptt mk` from container root (not inside any worktree) | Detect this case (cwd has `.bare/` but is not inside a worktree) and create worktree at `cwd/name` |
| `.pttconfig/` exists in both container root and main worktree | Container root takes precedence |
| User deletes main worktree in bare repo | `ptt cd` (no args) falls back to first non-bare worktree |
| Submodules during conversion | Warn: "submodules detected -- verify submodule paths after conversion" |
| Running from outside git repo after conversion | The `.git` pointer file in container root makes git recognize the directory |
| Branch with slashes (`feature/auth`) | Convert to directory-safe name (`feature-auth`) for worktree directory |

---

## Sources

**HIGH confidence (official docs, codebase analysis):**
- [Git worktree official documentation](https://git-scm.com/docs/git-worktree) -- main worktree vs linked worktree semantics, GIT_DIR/GIT_COMMON_DIR
- [Git config official documentation](https://git-scm.com/docs/git-config) -- worktree config scope, bare repo config
- ptt codebase analysis (`repo.go`, `worktree.go`, `resolve.go`, `config/resolve.go`)

**MEDIUM confidence (verified across multiple sources):**
- [Morgan Cugerone: How to use git worktree and in a clean way](https://morgan.cugerone.com/blog/how-to-use-git-worktree-and-in-a-clean-way/) -- canonical `.bare/` + `.git` pointer pattern
- [Morgan Cugerone: Workarounds for bare repo fetch issues](https://morgan.cugerone.com/blog/workarounds-to-git-worktree-using-bare-repository-and-cannot-fetch-remote-branches/) -- `remote.origin.fetch` fix
- [Andreas Schneider: Sliced bread: git-worktree and bare repo](https://blog.cryptomilk.org/2023/02/10/sliced-bread-git-worktree-and-bare-repo/) -- bare repo setup, git aliases
- [Nick Nisi: How I use git worktrees](https://nicknisi.com/posts/git-worktrees/) -- real-world bare repo usage, `git-bare-clone` script
- [Pablo Arias: Exploring Git Worktrees](https://pabloariasal.github.io/2023/12/27/git-worktrees/) -- bare clone workflow
- [Thomas Frans: A gentle introduction to Git worktree](https://gist.github.com/ThomasFrans/ab1cb531410ab0cd0616a88a735dd840) -- worktree.guessRemote, setup gotchas
- [matklad: How I Use Git Worktrees](https://matklad.github.io/2024/07/25/git-worktrees.html) -- fixed worktree set pattern (main/review/hotfix/work/scratch)
- [git-prole: announcing blog post](https://becca.ooo/blog/announcing-git-prole/) -- `convert` command design, bare repo layout

**MEDIUM confidence (tool READMEs, single source):**
- [git-wt by Ahmed El Gabri](https://gabri.me/blog/git-wt) -- migrate command, bare repo structure
- [worktree-manager (wtm)](https://github.com/jarredkenny/worktree-manager) -- post_create hooks, cleanup command
- [gwq](https://github.com/d-kuro/gwq) -- global worktree management, config hierarchy
- [wtp](https://github.com/satococoa/wtp) -- .wtp.yml config, copy/symlink/command hooks
- [worktrunk](https://github.com/max-sixty/worktrunk) -- AI agent workflows
- [git-worktree-runner (gtr)](https://github.com/coderabbitai/git-worktree-runner) -- .gtrconfig, editor integration
- [git-worktree-wrapper](https://github.com/lu0/git-worktree-wrapper) -- git checkout/branch wrapping approach

**LOW confidence (needs verification during implementation):**
- Exact behavior of `git rev-parse --show-toplevel` when run from inside a worktree nested in a bare repo container (needs integration testing)
- Whether `git worktree add` from inside a nested worktree correctly resolves relative paths to container root
- Submodule behavior during bare repo conversion
