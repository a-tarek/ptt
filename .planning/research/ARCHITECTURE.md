# Architecture: Bare Repo Integration

**Domain:** Bare repo conversion + nested worktree support for ptt CLI
**Researched:** 2026-02-09
**Overall confidence:** HIGH

---

## Executive Summary

The bare repo features (`mk-bare-repo`, nested `mk`, `cd` rename, config resolution changes) integrate cleanly with the existing ptt architecture. The codebase already contains partial bare repo awareness -- `GetRepoRoot()`, `GetHomePath()`, and `WorktreePath()` in `internal/git/repo.go` already detect bare repos and switch behavior. The integration requires one new command file, surgical modifications to five existing packages, and shell wrapper updates. No new internal packages are needed.

---

## 1. Where Does Bare Repo Detection Logic Live?

### Current State

Bare repo detection already exists in `internal/git/repo.go`:

- **`IsBareRepository()`** (line 12) -- Shells out to `git rev-parse --is-bare-repository`, returns bool.
- **`GetRepoRoot()`** (line 29) -- Already branches on bare/non-bare:
  - Bare: uses `git rev-parse --git-dir` (returns bare repo root)
  - Non-bare: uses `git rev-parse --show-toplevel` (returns checkout root)
- **`GetHomePath()`** (line 62) -- Already parses `git worktree list --porcelain` to find bare entries and match HEAD branch to a worktree.
- **`WorktreePath()`** (line 168) -- Already has bare/non-bare branching:
  - Bare: nested mode (`filepath.Dir(repoRoot)/name`)
  - Non-bare: sibling mode (`parentDir/repoName-name`)

### Required Changes

The existing detection logic is mostly correct but has issues:

**Problem 1: `WorktreePath()` bare detection is fragile (line 174)**

The current code checks `strings.HasSuffix(repoRoot, ".git")` as a heuristic, then falls back to parsing `git worktree list --porcelain` and stopping at the first empty line (only checks the first worktree entry). This is unreliable because:
- Bare repos do not necessarily end in `.git` (e.g., ptt creates them as `project-bare/`)
- The first entry may not be the bare one in all git versions

**Fix:** Replace with a single call to `IsBareRepository()` which is already correct and authoritative. The function already exists; `WorktreePath` just does not use it.

**Problem 2: `WorktreePath()` nested path calculation is wrong for ptt's model (line 200)**

Currently:
```go
parentDir := filepath.Dir(repoRoot)
targetPath = filepath.Join(parentDir, name)
```

For the ptt bare repo model (`project-bare/` contains bare git data + worktrees), `repoRoot` from `GetRepoRoot()` returns the bare repo root itself. We need worktrees nested **inside** that directory, not as siblings of it. The correct path is:

```go
targetPath = filepath.Join(repoRoot, name)
```

Wait -- this depends on the bare repo structure ptt creates. Let me trace this carefully.

### Bare Repo Structure (ptt model)

```
project-bare/                 <-- parent directory
  .bare/                      <-- actual git database (git clone --bare ... .bare)
  .git                        <-- file containing "gitdir: ./.bare"
  .pttconfig/                 <-- config, shared by all worktrees
    default
  main/                       <-- worktree for main branch
  feature-x/                  <-- worktree for feature-x branch
```

Key insight: `git rev-parse --git-dir` when run inside a worktree under this structure returns `project-bare/.bare/worktrees/<id>` (for linked worktrees) or `project-bare/.bare` (from inside the bare repo root). `git rev-parse --git-common-dir` returns `project-bare/.bare` in all cases.

For ptt, the "bare repo root" that matters is `project-bare/` (the parent containing `.bare/`, `.git`, and worktrees), NOT `.bare/` itself.

**Detection strategy:** When inside a worktree under a bare repo:
1. `git rev-parse --git-common-dir` returns path to `.bare/` directory
2. `filepath.Dir(commonDir)` gives us `project-bare/` -- the ptt bare repo root
3. New worktrees go at `project-bare/<name>`

### Recommended Architecture

Add a new function to `internal/git/repo.go`:

```go
// BareRepoRoot returns the ptt bare repo root directory.
// This is the parent of .bare/ -- the directory containing worktrees and .pttconfig/.
// Returns ("", false, nil) if not inside a bare repo structure.
func BareRepoRoot() (string, bool, error)
```

This function:
1. Calls `IsBareRepository()` or checks `git rev-parse --git-common-dir`
2. Determines if the repo follows ptt's bare structure (has `.bare/` and `.git` file)
3. Returns the parent directory of `.bare/` as the root

Then update `WorktreePath()` to use `BareRepoRoot()` instead of its current heuristics.

### Confidence: HIGH

Based on direct code reading of `internal/git/repo.go` and verified git rev-parse behavior from official docs.

---

## 2. How Does `mk-bare-repo` Interact with Existing Packages?

### New Command: `cmd/mk_bare_repo.go`

This is an entirely new cobra command. It does NOT modify any existing command.

**Flow:**

```
User runs: ptt mk-bare-repo [--name <dir-name>]

1. Validate: must be inside a non-bare git repo (error if already bare)
2. Validate: must have a remote origin (need URL for clone)
3. Determine source info:
   - remote URL: git remote get-url origin
   - current branch: git branch --show-current
   - repo name: filepath.Base(currentRoot)
4. Compute target: <parent>/<repo>-bare/  (or --name override)
5. Validate target does not exist
6. Execute conversion:
   a. mkdir <target>
   b. git clone --bare <remote-url> <target>/.bare
   c. echo "gitdir: ./.bare" > <target>/.git
   d. git -C <target> config remote.origin.fetch "+refs/heads/*:refs/remotes/origin/*"
   e. git -C <target> fetch origin
   f. git -C <target> worktree add main  (or whatever the default branch is)
7. Copy .pttconfig/ from source to target (if exists)
8. Output path for shell cd (via --output-path protocol)
```

### Package Interactions

| Package | Interaction | New/Modified |
|---------|-------------|--------------|
| `internal/git/repo.go` | `GetRepoRoot()`, `IsBareRepository()`, `CurrentBranch()` for pre-validation | **Existing** (no changes) |
| `internal/git/` | New helper: `GetRemoteURL()` to get `git remote get-url origin` | **New function** |
| `internal/git/` | New helper: `GetDefaultBranch()` to detect main/master | **New function** |
| `internal/config/resolve.go` | Used to check if source has `.pttconfig/` to copy | **Existing** (no changes) |
| `internal/setup/copy.go` | `CopyPath()` to copy `.pttconfig/` directory | **Existing** (no changes) |
| `internal/ui/task.go` | `TaskList` for progress display | **Existing** (no changes) |
| `cmd/` | New file `mk_bare_repo.go` | **New file** |

### Key Design Decisions

**Clone from remote, not local restructure:** The command runs `git clone --bare <remote-url>` rather than trying to restructure the `.git/` directory in place. This is:
- Safer (original repo is untouched)
- Simpler (no manual git internals manipulation)
- Idempotent (user can retry after failure)
- Consistent with the community pattern (Morgan Cugerone's approach)

**Remote URL requirement:** The command needs a remote to clone from. A purely local repo with no remote cannot be converted this way. This is an acceptable constraint because bare repos are most useful for remote-tracked repos.

**`.pttconfig/` copy:** If the source repo has `.pttconfig/`, it gets copied to the bare repo root. This is a simple filesystem copy, not a git-tracked operation.

### Confidence: HIGH

The conversion flow uses only standard git commands. The package interactions are minimal and well-understood.

---

## 3. How Does `mk` Change Behavior Inside a Bare Repo?

### Current `mk` Flow (cmd/new.go)

```
1. git.GetHomePath()           -> homePath (source for config resolution)
2. git.CurrentWorktreeRoot()   -> srcRoot (source for copy/symlink)
3. git.WorktreePath(homePath, name) -> targetPath
4. config.ResolveConfigPath(homePath, ...)  -> config file
5. config.ValidateActions(currentWorktreeRoot, actions) -> validate sources exist
6. git worktree add <targetPath> -b <branchName>
7. setup.ExecuteActions(currentWorktreeRoot, targetPath, ...)
8. Output targetPath
```

### What Changes for Bare Repos

The `mk` command itself needs **minimal changes** because the branching logic is already handled by `WorktreePath()` and `GetHomePath()`. Here is what changes:

**A. Path computation (already partially handled):**

`WorktreePath()` already switches to nested mode for bare repos. It just needs the fix described in section 1 (use `BareRepoRoot()` for correct root detection and path computation).

- Non-bare: `<parent>/repo-name` -> target `<parent>/repo-name-feature`
- Bare: `project-bare/` -> target `project-bare/feature`

**B. Config resolution root:**

Currently `mk` calls `config.ResolveConfigPath(homePath, ...)`. In a bare repo:
- `homePath` comes from `GetHomePath()` which returns the main worktree path (e.g., `project-bare/main/`)
- But `.pttconfig/` lives at the **bare repo root** (e.g., `project-bare/`), NOT inside a worktree

This is the critical change. Config resolution needs to look at the bare repo root, not the home worktree.

**C. Copy/symlink source root:**

Currently `mk` uses `CurrentWorktreeRoot()` as the source for copy/symlink actions. In a bare repo context, this is still correct -- you copy from whatever worktree you are currently in.

**D. Worktree creation command:**

`git worktree add <path> -b <branch>` works the same in bare and non-bare repos. No change needed.

### Modified Logic in `cmd/new.go`

```go
// Before (current):
configPath, err = config.ResolveConfigPath(homePath, configFlag)

// After:
configRoot := homePath
if bareRoot, isBare, err := git.BareRepoRoot(); err == nil && isBare {
    configRoot = bareRoot
}
configPath, err = config.ResolveConfigPath(configRoot, configFlag)
```

The same pattern applies to `cmd/eject.go` which has identical config resolution logic.

### Confidence: HIGH

Direct code analysis of `cmd/new.go` lines 37-49 and 62-85. The change is surgical.

---

## 4. How Does Config Resolution Change?

### Current Config Resolution (`internal/config/resolve.go`)

```go
func ResolveConfigPath(repoRoot string, name string) (string, error)
```

Takes `repoRoot` and looks for:
- Empty name: `<repoRoot>/.pttconfig/default` or `<repoRoot>/.wtconfig`
- Bare name: `<repoRoot>/.pttconfig/<name>` or `<repoRoot>/.wtconfig-<name>`
- Path with `/`: exact path

### What Changes

**The function itself does NOT change.** The resolution logic is correct -- it just needs to receive the right `repoRoot`.

The change is in **callers** -- they need to pass the bare repo root instead of the home worktree path when in a bare repo context.

### Affected Callers

| File | Current `repoRoot` Source | Bare Repo `repoRoot` |
|------|---------------------------|----------------------|
| `cmd/new.go` (line 66-67) | `homePath` (from `GetHomePath()`) | `BareRepoRoot()` |
| `cmd/eject.go` (line 182) | `srcRoot` (from `CurrentWorktreeRoot()`) | `BareRepoRoot()` |
| `cmd/init_cmd.go` (line 43-49) | `repoRoot` (from `GetHomePath()`) | `BareRepoRoot()` |

### Config Layout in Bare Repos

```
project-bare/                   <-- BareRepoRoot()
  .bare/
  .git
  .pttconfig/                   <-- Config lives HERE
    default
    ci
  main/                         <-- GetHomePath() returns this
    .pttconfig/                 <-- Does NOT exist here
    src/
  feature-x/
    src/
```

### Validation Source Root

`config.ValidateActions(srcRoot, actions)` checks that source files exist at `<srcRoot>/<path>`. In a bare repo, `srcRoot` should remain the **current worktree root** (where files actually are), NOT the bare repo root. This is already correct in the current code.

### `ptt init` Changes

`ptt init` creates `.pttconfig/default` at `GetHomePath()`. In a bare repo, it should create it at `BareRepoRoot()` instead. Same pattern as `mk` -- detect bare repo, use bare root for config location.

### Recommended Helper

Add a convenience function to centralize config root determination:

```go
// internal/git/repo.go
func ConfigRoot() (string, error) {
    if bareRoot, isBare, err := BareRepoRoot(); err == nil && isBare {
        return bareRoot, nil
    }
    return GetHomePath()
}
```

Then all callers use `git.ConfigRoot()` instead of `git.GetHomePath()` for config resolution.

### Confidence: HIGH

Direct code analysis. The separation between "config root" and "source root" is clean and well-bounded.

---

## 5. How Does the `go` to `cd` Rename Affect Shell Wrappers?

### Current Shell Wrapper (all three shells)

```bash
# wrapper.zsh (and .bash, .fish equivalently)
ptt() {
  case "$1" in
    go|goto|home|mk|new|eject)
      # ... capture --output-path, cd to result
      ;;
    *)
      "__PTT_BIN__" "$@"
      ;;
  esac
}
```

### Changes Required

**A. Add `cd` to the case list:**

```bash
case "$1" in
    cd|go|goto|home|mk|new|eject)
```

This is a one-line change in each of the three wrapper templates:
- `internal/shell/templates/wrapper.bash`
- `internal/shell/templates/wrapper.zsh`
- `internal/shell/templates/wrapper.fish`

**B. Rename the cobra command in `cmd/goto.go`:**

```go
// Before:
Use: "go [worktree]",
Aliases: []string{"goto", "home"},

// After:
Use: "cd [worktree]",
Aliases: []string{"go", "goto", "home"},
```

`go` becomes an alias for backward compatibility. The file could also be renamed from `goto.go` to `cd.go` for clarity, though this is cosmetic.

**C. Update `cmd/root.go` help text:**

```go
// Before:
"  go [worktree]      Navigate to a worktree (or home)"

// After:
"  cd [worktree]      Navigate to a worktree (or home)"
```

**D. Update completions:**

Cobra completions are auto-generated from command definitions. Renaming the command `Use` field automatically updates completions. No separate completion code changes needed.

**E. Shell function name conflict:**

`cd` is a shell builtin. However, there is no conflict because:
- The shell wrapper function is named `ptt()`, not `cd()`
- `cd` is a **subcommand** of `ptt`, not a standalone command
- `ptt cd foo` calls the binary which outputs a path, then the wrapper runs the real `cd`

This is safe. The wrapper calls `cd "$result"` internally, which invokes the builtin `cd`, not a recursive call.

### Confidence: HIGH

Direct code analysis of all three wrapper templates and `cmd/goto.go`.

---

## Component Map: New vs Modified

### New Components

| Component | File | Purpose |
|-----------|------|---------|
| `mk-bare-repo` command | `cmd/mk_bare_repo.go` | New cobra command for bare repo conversion |
| `GetRemoteURL()` | `internal/git/repo.go` | Get `origin` remote URL |
| `GetDefaultBranch()` | `internal/git/repo.go` | Detect main/master as default branch |
| `BareRepoRoot()` | `internal/git/repo.go` | Detect ptt bare repo structure, return root |
| `ConfigRoot()` | `internal/git/repo.go` | Convenience: bare root or home path for config |

### Modified Components

| Component | File | Change |
|-----------|------|--------|
| `cd` command | `cmd/goto.go` | Rename `Use` from `go` to `cd`, add `go` as alias |
| `mk` command | `cmd/new.go` | Use `ConfigRoot()` for config resolution |
| `eject` command | `cmd/eject.go` | Use `ConfigRoot()` for config resolution |
| `init` command | `cmd/init_cmd.go` | Use `ConfigRoot()` for `.pttconfig/` creation path |
| `WorktreePath()` | `internal/git/repo.go` | Fix bare repo detection and path calculation |
| Root help text | `cmd/root.go` | Update command listing |
| Shell wrappers | `internal/shell/templates/wrapper.{bash,zsh,fish}` | Add `cd` to case list |
| `ls` command | `cmd/list.go` | Filter out bare repo entry from display (optional) |

### Unchanged Components

| Component | Why Unchanged |
|-----------|---------------|
| `internal/config/resolve.go` | Logic is correct; callers pass different root |
| `internal/config/parser.go` | No changes needed |
| `internal/config/action.go` | No changes needed |
| `internal/config/validator.go` | No changes needed |
| `internal/config/flags.go` | No changes needed |
| `internal/setup/executor.go` | No changes needed |
| `internal/setup/copy.go` | Reused by `mk-bare-repo` for `.pttconfig/` copy |
| `internal/setup/symlink.go` | No changes needed |
| `internal/setup/run.go` | No changes needed |
| `internal/shell/detect.go` | No changes needed |
| `internal/shell/embed.go` | No changes needed |
| `internal/installer/` | No changes needed |
| `internal/ui/` | No changes needed |
| `cmd/completion.go` | No changes needed (auto-generated from cobra) |
| `cmd/merge.go` | No changes needed |
| `cmd/rebase.go` | No changes needed |
| `cmd/delete.go` | No changes needed |

---

## Data Flow Diagrams

### `ptt mk-bare-repo` Flow

```
[User runs ptt mk-bare-repo]
  |
  v
[Validate: not bare, has remote]
  |-- git.IsBareRepository() -> must be false
  |-- git.GetRemoteURL() -> must have origin
  |-- git.GetRepoRoot() -> sourceRoot
  |-- git.GetDefaultBranch() -> defaultBranch (main/master)
  |
  v
[Compute target path]
  |-- targetDir = filepath.Dir(sourceRoot) / filepath.Base(sourceRoot) + "-bare"
  |
  v
[Execute conversion]  (sequence of exec.Command calls)
  |-- mkdir targetDir
  |-- git clone --bare <remoteURL> targetDir/.bare
  |-- write "gitdir: ./.bare" to targetDir/.git
  |-- git -C targetDir config remote.origin.fetch "+refs/heads/*:refs/remotes/origin/*"
  |-- git -C targetDir fetch origin
  |-- git -C targetDir/.bare worktree add targetDir/<defaultBranch> <defaultBranch>
  |
  v
[Copy .pttconfig/ if exists]
  |-- setup.CopyPath(sourceRoot/.pttconfig, targetDir/.pttconfig)
  |
  v
[Output path]
  |-- fmt.Println(targetDir/<defaultBranch>)  // for shell wrapper cd
```

### `ptt mk <name>` in Bare Repo

```
[User runs ptt mk feature-x (inside project-bare/main/)]
  |
  v
[Resolve roots]
  |-- git.GetHomePath() -> project-bare/main/     (source for copy/symlink)
  |-- git.ConfigRoot()  -> project-bare/           (for .pttconfig/ resolution)
  |-- git.WorktreePath(homePath, "feature-x")
  |     |-- BareRepoRoot() -> project-bare/
  |     |-- return project-bare/feature-x
  |
  v
[Resolve config]
  |-- config.ResolveConfigPath(configRoot, "")
  |     -> project-bare/.pttconfig/default
  |
  v
[Validate sources against current worktree]
  |-- config.ValidateActions(currentWorktreeRoot, actions)
  |     -> checks project-bare/main/<path> exists
  |
  v
[Create worktree]
  |-- git worktree add project-bare/feature-x -b feature-x
  |
  v
[Execute config actions]
  |-- setup.ExecuteActions(currentWorktreeRoot, targetPath, actions, ...)
  |     -> copies from project-bare/main/ to project-bare/feature-x/
  |
  v
[Output path]
  |-- fmt.Println(project-bare/feature-x)
```

### `ptt cd <name>` Flow (renamed from `go`)

```
[Shell wrapper intercepts "cd" subcommand]
  |-- ptt --output-path cd feature-x
  |
  v
[cobra routes to cdCmd (was goCmd)]
  |-- git.ResolveWorktree("feature-x")
  |     -> matches worktree by basename
  |     -> returns Worktree{Path: "project-bare/feature-x", Branch: "feature-x"}
  |
  v
[Output path to stdout]
  |-- fmt.Println(wt.Path)
  |
  v
[Shell wrapper does: cd "project-bare/feature-x"]
```

---

## Bare Repo Root Detection: Deep Dive

This is the trickiest part of the integration. Here is the precise algorithm for `BareRepoRoot()`:

```go
func BareRepoRoot() (string, bool, error) {
    // Step 1: Get the common git dir (shared across all worktrees)
    cmd := exec.Command("git", "rev-parse", "--git-common-dir")
    output, err := cmd.Output()
    if err != nil {
        return "", false, err
    }
    commonDir := strings.TrimSpace(string(output))

    // Step 2: Make absolute
    if !filepath.IsAbs(commonDir) {
        cwd, _ := os.Getwd()
        commonDir = filepath.Join(cwd, commonDir)
    }
    commonDir = filepath.Clean(commonDir)

    // Step 3: Check if this is a ptt bare repo structure
    // In ptt's model, commonDir is project-bare/.bare
    // The parent should have a .git file (not directory) containing "gitdir: ./.bare"
    parentDir := filepath.Dir(commonDir)
    gitFilePath := filepath.Join(parentDir, ".git")

    info, err := os.Lstat(gitFilePath)
    if err != nil {
        return "", false, nil  // Not a ptt bare repo structure
    }

    // Must be a file (not a directory) -- gitdir pointer
    if info.IsDir() {
        return "", false, nil  // Regular repo, not bare
    }

    return parentDir, true, nil
}
```

**Why this works:**
- In a regular repo, `--git-common-dir` returns `.git` (a directory). Its parent has `.git` as a directory, so the `Lstat` check fails (it IS a directory).
- In a ptt bare repo, from any worktree, `--git-common-dir` returns `project-bare/.bare`. The parent `project-bare/` has a `.git` **file** pointing to `.bare/`. This is the signal.
- This handles being inside any worktree (main, feature-x, etc.) because `--git-common-dir` always resolves to the shared `.bare/`.

**Edge case:** What if the user is running from inside the `project-bare/` directory itself (not inside any worktree)? In that case, `git rev-parse --git-common-dir` still works because the `.git` file in that directory points git to `.bare/`.

### Confidence: HIGH

Verified against git rev-parse documentation. The `.git`-file-not-directory heuristic is reliable because it is the mechanism git itself uses to link worktrees.

---

## `WorktreePath()` Fix for Nested Mode

Current broken logic (line 198-201):
```go
if isBare {
    parentDir := filepath.Dir(repoRoot)
    targetPath = filepath.Join(parentDir, name)
}
```

The problem: `repoRoot` here comes from `GetHomePath()` which returns the home **worktree** path (e.g., `project-bare/main/`). So `filepath.Dir()` gives `project-bare/` and `filepath.Join(project-bare/, name)` gives `project-bare/name`. This **accidentally works** for the ptt model but for the wrong reason.

However, if `repoRoot` were changed to come from `GetRepoRoot()` (as the parameter name suggests), it would return `project-bare/.bare` for bare repos, and then `filepath.Dir()` gives `project-bare/` which also works.

**Recommendation:** Make `WorktreePath()` explicitly use `BareRepoRoot()`:

```go
func WorktreePath(repoRoot string, name string) (string, error) {
    bareRoot, isBare, err := BareRepoRoot()
    if err != nil {
        return "", err
    }

    var targetPath string
    if isBare {
        targetPath = filepath.Join(bareRoot, name)
    } else {
        parentDir := filepath.Dir(repoRoot)
        repoName := filepath.Base(repoRoot)
        targetPath = filepath.Join(parentDir, repoName+"-"+name)
    }

    if _, err := os.Stat(targetPath); err == nil {
        return "", fmt.Errorf("path already exists: %s", targetPath)
    }

    return targetPath, nil
}
```

---

## `ResolveWorktree()` Changes for Bare Repos

The existing `ResolveWorktree()` in `internal/git/resolve.go` matches worktrees by basename using suffix matching: `basename == name || strings.HasSuffix(basename, "-"+name)`.

In bare repos, worktree basenames are just the name itself (e.g., `feature-x` not `repo-feature-x`). The exact match `basename == name` already handles this. The suffix match (`-name`) may produce false positives if worktree names contain dashes, but this is the same risk as in non-bare repos. No change needed.

### Confidence: HIGH

---

## `ls` Command Changes for Bare Repos

The `ls` command currently lists ALL worktrees from `git worktree list --porcelain`, including the bare repo entry itself. In a bare repo context, the bare entry shows up as:

```
  project-bare/.bare             (bare)
```

This is noise -- users care about their worktrees, not the bare repo metadata directory.

**Recommended fix:** Filter out entries where `wt.IsBare == true` in `cmd/list.go`:

```go
for _, wt := range worktrees {
    if wt.IsBare {
        continue  // Skip bare repo entry
    }
    // ... display logic
}
```

### Confidence: HIGH

---

## Suggested Build Order

Based on dependency analysis:

### Phase A: Foundation (no user-visible changes)

1. **Add `BareRepoRoot()` and `ConfigRoot()` to `internal/git/repo.go`**
   - Pure additions, no existing behavior changes
   - Write tests using real bare repo setup in temp dirs
   - This unblocks everything else

2. **Fix `WorktreePath()` bare detection**
   - Replace heuristic with `BareRepoRoot()` call
   - Update existing test `TestWorktreePath_BareRepo` (currently skipped)

### Phase B: `cd` rename (smallest, most independent change)

3. **Rename `go` to `cd` in `cmd/goto.go`**
   - Change `Use`, add `go` as alias
   - Update `cmd/root.go` help text
   - Optionally rename file to `cmd/cd.go`

4. **Update shell wrappers**
   - Add `cd` to case list in all three templates
   - Existing aliases (`go`, `goto`, `home`) remain in case list for backward compat

### Phase C: Config resolution (needed before `mk-bare-repo` and bare `mk`)

5. **Update config resolution callers to use `ConfigRoot()`**
   - `cmd/new.go`: use `ConfigRoot()` for config path, keep `CurrentWorktreeRoot()` for source
   - `cmd/eject.go`: same pattern
   - `cmd/init_cmd.go`: use `ConfigRoot()` for `.pttconfig/` creation

### Phase D: `mk-bare-repo` command (the big new feature)

6. **Add `GetRemoteURL()` and `GetDefaultBranch()` helpers**
   - Simple git command wrappers

7. **Implement `cmd/mk_bare_repo.go`**
   - Full conversion flow
   - Integration test with real git repos

### Phase E: Polish

8. **Update `ls` to filter bare entries**
9. **Update README / docs**
10. **End-to-end testing of full bare repo workflow**

### Dependency Graph

```
Phase A: BareRepoRoot(), ConfigRoot(), WorktreePath fix
  |
  +---> Phase B: cd rename (independent of A, but ordered after for cleanliness)
  |
  +---> Phase C: Config resolution callers (depends on A)
  |       |
  |       +---> Phase D: mk-bare-repo (depends on A + C)
  |
  +---> Phase E: ls filter, docs (depends on all above)
```

**Phase B is fully independent** and could be done first or in parallel with Phase A.

---

## Anti-Patterns to Avoid

### 1. Do NOT put bare repo logic in `internal/config/`

The config package should remain unaware of bare repos. Config resolution is just "given a root path, find the config file." The bare repo awareness belongs in `internal/git/` and the callers.

### 2. Do NOT create a new `internal/bare/` package

The bare repo logic is small (one detection function + two helpers). It belongs in `internal/git/repo.go` alongside the existing `IsBareRepository()`, `GetRepoRoot()`, and `WorktreePath()`.

### 3. Do NOT restructure the existing repo in-place

The `mk-bare-repo` command creates a **new** bare repo directory alongside the original. Users verify it works, then manually delete the old repo. Never modify or delete the user's existing repository.

### 4. Do NOT change `ResolveConfigPath()` signature

The function is correct as-is. Changing its interface to accept a "repo type" parameter would leak bare repo awareness into the config package. Instead, callers compute the correct root and pass it in.

### 5. Do NOT use `git init --bare` for conversion

`git clone --bare <remote-url>` is safer than `git init --bare` + manual copying because it handles all git internals correctly (packfiles, refs, hooks, etc.).

---

## Sources

- [Git rev-parse documentation](https://git-scm.com/docs/git-rev-parse) -- `--git-common-dir`, `--is-bare-repository`, `--git-dir` behavior (HIGH confidence)
- [Git worktree documentation](https://git-scm.com/docs/git-worktree) -- worktree list, bare repo display, linked worktree mechanics (HIGH confidence)
- [Git clone documentation](https://git-scm.com/docs/git-clone) -- `--bare` flag behavior and limitations (HIGH confidence)
- [Morgan Cugerone: How to use git worktree in a clean way](https://morgan.cugerone.com/blog/how-to-use-git-worktree-and-in-a-clean-way/) -- `.bare/` + `.git` file pattern (MEDIUM confidence, community pattern)
- [Morgan Cugerone: Workarounds for bare repo fetch](https://morgan.cugerone.com/blog/workarounds-to-git-worktree-using-bare-repository-and-cannot-fetch-remote-branches/) -- `remote.origin.fetch` fix (MEDIUM confidence, verified against git docs)
- [Andreas Schneider: git-worktree and bare repo](https://blog.cryptomilk.org/2023/02/10/sliced-bread-git-worktree-and-bare-repo/) -- nested worktree pattern (MEDIUM confidence)
- Direct source code analysis of `internal/git/repo.go`, `cmd/new.go`, `cmd/goto.go`, shell wrapper templates (HIGH confidence)
