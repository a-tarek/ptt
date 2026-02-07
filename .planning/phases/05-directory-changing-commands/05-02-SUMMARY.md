---
phase: 05
plan: 02
subsystem: directory-commands
tags: [worktree-creation, config-integration, placement-logic]
requires: [05-01, 04-02]
provides:
  - wt new command with config integration
  - automatic worktree placement (bare=nested, regular=sibling)
  - branch creation and reuse logic
affects: [05-03]
tech-stack:
  added: []
  patterns: [config-merge, automatic-rollback, stderr-confirmations]
key-files:
  created:
    - cmd/new.go
    - cmd/new_test.go
  modified:
    - internal/git/repo.go
decisions:
  - slug: config-merge-on-top
    desc: "--copy/--symlink flags merge with file-based config, apply independently of --skip-config"
    rationale: "Allows users to skip .wtconfig but still apply inline overrides"
    date: 2026-02-07
  - slug: silent-config-skip
    desc: "Missing .wtconfig is silently skipped, not an error"
    rationale: ".wtconfig is optional - only projects that need it create it"
    date: 2026-02-07
  - slug: bare-detection-via-home-path
    desc: "WorktreePath checks if home path is bare, not current directory"
    rationale: "When in a worktree of bare repo, IsBareRepository() returns false"
    date: 2026-02-07
metrics:
  duration: 2m
  completed: 2026-02-07
---

# Phase 5 Plan 2: wt new Command Summary

**One-liner:** wt new creates worktrees with automatic placement, config integration, and branch handling

## What Was Built

Implemented the `wt new` command, the most-used command in the tool. Creates git worktrees with:
- Automatic placement (nested for bare repos, sibling for regular repos)
- Branch creation (new from HEAD) or branch reuse (existing branch)
- Config integration (.wtconfig by default, --config for variants, --skip-config to bypass)
- Inline flag overrides (--copy/--symlink merge with config)
- Automatic rollback on config action failure
- --output-path protocol for shell wrapper coordination

## Task Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1. Implement wt new command | 6259791 | feat(05-02): implement wt new command |
| 2. Test wt new integration | 91dc00e | test(05-02): add integration tests for wt new command |

## Technical Details

### Architecture

```
wt new <name>
  ├─ git.GetHomePath() → home worktree path
  ├─ git.WorktreePath(home, name) → target path (bare=nested, regular=sibling)
  ├─ git worktree add <target> -b <branch> → create worktree + new branch
  │   └─ fallback: git worktree add <target> <branch> → existing branch
  ├─ config.ResolveConfigPath(home, flag) → .wtconfig or .wtconfig-<name>
  ├─ config.ParseFile(path) → actions
  ├─ config.ValidateActions(srcRoot, actions) → check sources exist
  ├─ config.BuildActionsFromFlags(copy, symlink, nil) → inline actions
  ├─ setup.ExecuteActions(src, target, merged) → apply config (rollback on failure)
  └─ git.IsDirty(target) → check status for confirmation
```

### Flag Hierarchy

- Default: load `.wtconfig` if it exists (silent skip if absent)
- `--config ci`: load `.wtconfig-ci` instead (errors if not found)
- `--skip-config`: skip file-based config only
- `--copy`/`--symlink`: ALWAYS apply (merge on top, independent of --skip-config)

### Worktree Placement

**Bare repository** (repo.git):
```
repo.git/
  ├─ main/        (first worktree)
  ├─ staging/     (wt new staging)
  └─ feature/     (wt new feature)
```

**Regular repository** (repo):
```
parent/
  ├─ repo/           (main checkout)
  ├─ repo-staging/   (wt new staging)
  └─ repo-feature/   (wt new feature)
```

### Output Protocol

**stderr**: All confirmation messages
- "Created worktree {name} (branch: {branch})"
- "Applied {config} ({N} actions)" (if config applied)
- "Switched to {name} (branch: {branch}, {clean|dirty})"

**stdout**:
- If `--output-path`: only the target path (for shell wrapper cd)
- If not: duplicate "Switched to..." message (for user when running manually)

## Decisions Made

### 1. Config Merge on Top
**Decision:** `--copy`/`--symlink` flags merge with file-based config and apply independently of `--skip-config`

**Rationale:** Allows users to skip .wtconfig but still apply inline overrides. Common pattern: `wt new --skip-config --copy .env staging`

**Impact:** Flag actions append after config actions, no duplicate check across sources

### 2. Silent Config Skip
**Decision:** Missing .wtconfig is silently skipped, not an error

**Rationale:** .wtconfig is optional - only projects that need it create it. Making it required would force all projects to have empty config files.

**Impact:** Only error if `--config <name>` specified and file not found

### 3. Bare Detection via Home Path
**Decision:** WorktreePath checks if home path is bare, not current directory

**Rationale:** When in a worktree of a bare repo, `IsBareRepository()` returns false because the worktree itself isn't bare. Need to check the home worktree properties instead.

**Implementation:** Check for .git suffix or look for 'bare' marker in first entry of `git worktree list --porcelain`

**Impact:** Fixes incorrect sibling placement when running from bare repo worktree

## Test Coverage

Integration tests with real git repositories:
- **TestNewCommandCreatesWorktree**: Basic worktree creation, branch creation
- **TestNewCommandExistingBranch**: Existing branch reuse
- **TestNewCommandWithSkipConfig**: --skip-config flag honors config bypass
- **TestNewCommandAlreadyExists**: Duplicate path error handling

Test helper `setupBareRepo` creates temporary bare repo with initial worktree for realistic testing.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed bare repo detection in WorktreePath**
- **Found during:** Task 2 test execution
- **Issue:** Tests were failing because WorktreePath was checking if current directory is bare (returns false when in worktree of bare repo), causing sibling placement instead of nested
- **Fix:** Check home path properties instead - look for .git suffix or 'bare' marker in worktree list
- **Files modified:** internal/git/repo.go
- **Commit:** 91dc00e (included in test commit)

## Files Modified

### Created
- **cmd/new.go** (162 lines): wt new command implementation with full config integration
- **cmd/new_test.go** (238 lines): Integration tests with setupBareRepo helper

### Modified
- **internal/git/repo.go**: Fixed WorktreePath bare detection (checks home path, not cwd)

## Next Phase Readiness

**Ready for 05-03 (eject command):**
- ✓ Worktree creation with placement logic working
- ✓ Config integration tested and verified
- ✓ --output-path protocol established
- ✓ Bare vs regular detection reliable

**Dependencies satisfied:**
- Phase 4 config system (ResolveConfigPath, ParseFile, ValidateActions, BuildActionsFromFlags)
- Phase 5.01 git helpers (GetHomePath, WorktreePath, CurrentWorktreeRoot, IsDirty)

## Self-Check: PASSED
