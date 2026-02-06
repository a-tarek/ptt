# Codebase Concerns

**Analysis Date:** 2026-02-07

## Tech Debt

**Subshell/Piping Data Loss in Helper Functions:**
- Issue: Functions `_wt_resolve_path()` and `_wt_resolve_branch()` use piped `while` loops (lines 450, 466), which execute in subshells. Return values from within subshells don't propagate to the caller, and local variable modifications are lost.
- Files: `wt.zsh` (lines 446-480)
- Impact: While the `echo` output is captured correctly, any variable state changes or return codes from within the subshell are discarded. This limits future refactoring and makes the code fragile.
- Fix approach: Replace piped `while` loops with `while` loops that read from process substitution (`while ... < <(...)`) or use array iteration. This keeps execution in the parent shell context.

**Duplicated Logic Across Helper Functions:**
- Issue: `_wt_resolve_path()` (lines 446-460) and `_wt_resolve_branch()` (lines 462-480) both parse `git worktree list` output with nearly identical matching logic. When the matching algorithm changes (e.g., worktree naming convention), both must be updated.
- Files: `wt.zsh` (lines 446-480)
- Impact: Maintenance burden increases with each feature that uses worktree resolution. Bug fixes in one function may not be applied to the other.
- Fix approach: Extract common worktree-matching logic into a single helper function that returns matching worktree entry and branch, then have `_wt_resolve_path()` and `_wt_resolve_branch()` call it to extract their respective values.

**Inconsistent Error Exit Codes:**
- Issue: Throughout the codebase, functions return `1` on error, but return codes from failed git commands are suppressed with `&>/dev/null` (multiple instances at lines 54, 72, 107, 117, 240, 251, 264). When operations fail silently, callers can't distinguish between different failure modes.
- Files: `wt.zsh` (lines 36-551)
- Impact: Debugging is harder when error messages are lost. Users see generic "Error: failed to..." messages without the underlying git error details.
- Fix approach: Capture stderr separately and log it on failure. Reserve `&>/dev/null` for expected failures only (like checking branch existence).

## Known Bugs

**Stash State Race Condition in `_wt_eject`:**
- Symptoms: If multiple `wt eject` commands run on the same repository simultaneously in different worktrees, stash counting (lines 229-232) can give false positives due to shared git stash state.
- Files: `wt.zsh` (lines 226-237)
- Trigger: Run `wt eject` in two worktrees at nearly the same time against the same repo
- Workaround: Ensure sequential execution of `wt eject`; don't parallelize across worktrees of the same repo

**Detached HEAD Stash Pop Failure Not Handled:**
- Symptoms: If stash pop fails during `_wt_eject` (line 264), the error is silenced but uncommitted changes are not restored to the user's working directory.
- Files: `wt.zsh` (lines 262-266)
- Trigger: Create a stash conflict scenario during eject (e.g., uncommitted changes conflict with restored branch state)
- Workaround: Manually inspect git stash list and pop the stash manually with conflict resolution

**Variable Scope Leak in Parsing Loop:**
- Symptoms: In `_wt_list()` (lines 288-302), variables `wt_entry` and `branch` persist across loop iterations in the subshell context. While `wt_entry=""` is explicitly reset (line 300), this relies on explicit cleanup that could be forgotten in future modifications.
- Files: `wt.zsh` (lines 278-310)
- Trigger: Any refactoring of the output parsing logic without clearing variables
- Workaround: Carefully review loop variable initialization on each iteration

**Reserved Variable Collision Avoided:**
- Resolution: Variable `path` renamed to `entry` in `_wt_setup` (line 393) to avoid zsh reserved word collision. The `path` variable in zsh is a special array tied to `$PATH` environment variable. Using it as a local variable name can clobber PATH lookups, causing "command not found" errors for external commands while builtins still work.
- Files: `wt.zsh` (lines 388-396)
- Prevention: Future code MUST avoid using `path` as a variable name. Use `entry`, `file_path`, or other alternatives.
- Context: This was a preventive fix after discovering zsh's special `path` variable behavior.

## Security Considerations

**Unvalidated Paths in File Operations:**
- Risk: `cp` and `ln -s` operations use paths constructed from .wtconfig and git-derived strings without validation. A maliciously crafted entry in .wtconfig containing special characters could lead to unintended file operations.
- Files: `wt.zsh` (lines 405-443)
- Current mitigation: Paths use quoted variables and come from git or user-controlled .wtconfig
- Recommendations: Add explicit path validation before `cp` and `ln -s` operations. Consider using safer alternatives like `install` for copying or absolute path construction with normalization.

**Configurable File Operations via .wtconfig:**
- Risk: Files specified in .wtconfig are copied or symlinked based on user configuration. Secrets intended for a specific worktree context may leak to sibling worktrees if config specifies copy. Symlinked dependencies may cause conflicts between worktrees with different versions.
- Files: `wt.zsh` (lines 368-444), `.wtconfig` template (lines 115-144)
- Current mitigation: Users control .wtconfig contents; `--copy` and `--symlink` flags can override config
- Recommendations: Document risks in .wtconfig comments. Add validation that symlink targets are on same filesystem. Consider adding `--dry-run` to preview file operations before executing.

## Performance Bottlenecks

**`cp -r` on Large Directories:**
- Problem: Copying large directories recursively (line 407, 432) with `--copy` flag is extremely slow for large projects (> 10GB node_modules, .venv, etc.). No progress indicator or cancellation mechanism.
- Files: `wt.zsh` (lines 405-443)
- Cause: Shell `cp -r` has no parallel copy, no deduplication, and no resume on interruption
- Improvement path: Use `rsync -a --progress` instead of `cp -r`, or implement hardlink copying with `cp -al` (if filesystem supports it). Add `--quiet` flag to suppress output, and allow user to cancel with Ctrl+C.

**Repeated `git worktree list` Parsing:**
- Problem: Multiple functions call `git worktree list --porcelain` repeatedly (lines 107, 172, 177, 280, 450, 466, 491, 494). Each call spawns a git subprocess and parses output independently.
- Files: `wt.zsh` (lines 105-505)
- Cause: No caching of worktree metadata across function calls
- Improvement path: Cache `git worktree list` output in a variable at function entry if called multiple times in sequence, or refactor to build a lookup table once per command invocation.

**`git stash list | wc -l` Counts All Stashes:**
- Problem: Lines 229 and 232 count all stashes to detect if a new one was created. For repos with hundreds of stashes, this is O(n) and slow.
- Files: `wt.zsh` (lines 228-237)
- Cause: No built-in way to detect stash creation without counting
- Improvement path: Use `git stash create` to get the stash ID before and after, then check if IDs differ. Or use `git stash push` return code more reliably to detect success.

## Fragile Areas

**`_wt_eject` Multi-Step Transaction Without Rollback Guarantee:**
- Files: `wt.zsh` (lines 146-276)
- Why fragile: The eject operation involves 7+ git commands (stash, checkout, worktree add, stash pop). Partial failure at any step leaves the repo in an inconsistent state. Rollback attempts (lines 242-247, 253-258) may themselves fail or lose state.
- Safe modification: Do not change the order of stash/checkout/worktree/pop steps without understanding full state dependencies. Add comprehensive error logging to track which step failed. Consider wrapping the entire operation in a git transaction equivalent (not available in git, so add manual state snapshots before critical steps).
- Test coverage: Gaps in failure scenarios (e.g., what happens if stash pop fails after worktree is created). Test with intentional failures at each step.

**Worktree Name Resolution Ambiguity:**
- Files: `wt.zsh` (lines 446-505)
- Why fragile: The name matching (lines 454, 470, 498) uses suffix matching: `*"-${name}"` or exact match. A user naming a worktree "foo" could match repo "my-foo" unintentionally. Adding two worktrees with overlapping suffixes creates ambiguous resolution.
- Safe modification: Clarify naming constraints in documentation. Consider requiring exact match by default and adding a `--partial` flag for suffix matching.
- Test coverage: Gaps in multi-match scenarios. No tests for repos named similarly to worktree names.

**Fallback Branch Detection in `_wt_eject`:**
- Files: `wt.zsh` (lines 176-207)
- Why fragile: Fallback branch detection uses directory naming convention (line 193: `suffix="${dir_name#${repo_basename}-}"`) but this assumes consistent naming. A worktree created outside `wt` tool will not follow this pattern. The logic falls back to the full directory name as a branch name (line 196), which may not exist.
- Safe modification: Add validation that fallback branch actually exists before proceeding (already done at line 199). Add verbose logging of branch resolution choices.
- Test coverage: Gap in non-standard worktree naming created by git command directly.

**Symlink Breakage Across Filesystems:**
- Files: `wt.zsh` (lines 411-415, 437-441)
- Why fragile: Symlinked files/directories break if source and target worktrees are moved to different filesystems or machines. No validation that symlink target is still valid.
- Safe modification: Add a check after symlink creation that confirms the symlink resolves. Document that worktrees must stay on same filesystem as main repo.
- Test coverage: No tests for symlink validity after worktree operations.

**Override Precedence Complexity:**
- Files: `wt.zsh` (lines 368-444)
- Why fragile: The two-phase override application in `_wt_setup` (config file processing + one-off overrides) requires careful understanding. Phase 1 processes .wtconfig with CLI flag overrides taking precedence. Phase 2 applies CLI-only overrides not in config. If this logic changes, it's easy to double-process or miss overrides.
- Safe modification: Do not modify the two-phase application order without understanding the `applied` tracking mechanism. The associative array `applied` prevents double-processing of paths that appear in both config and CLI flags.
- Test coverage: Gaps in testing all combinations of config + CLI overrides.

## Scaling Limits

**Max Worktrees Per Repository:**
- Current capacity: Tested up to ~20 worktrees; likely hits filesystem limits or git performance issues
- Limit: Shell loops over all worktrees on each `wt` command (lines 450, 466, 494). With hundreds of worktrees, `git worktree list --porcelain` becomes slow (O(n) parsing).
- Scaling path: Implement worktree filtering by prefix or pattern early in pipeline. Cache worktree list globally. Consider git-worktree native speedups as they're released (git 2.36+).

**Large Directory Symlinks/Copies:**
- Current capacity: Works for typical projects (< 100K files in directories like node_modules, .venv, target)
- Limit: Symlinked directories break with monorepos having multiple independent dependency trees. Copy operations on large directories are slow.
- Scaling path: Support per-workspace symlink strategy. Add granular path specifications in .wtconfig. Consider using hardlinks for copy operations where possible.

## Dependencies at Risk

**Git Worktree Feature Maturity:**
- Risk: Git worktree is relatively young feature. Breaking changes or behavior shifts between git versions (especially 2.35 - 2.40 range) could break the tool.
- Impact: Tool stops working with newer git versions that change worktree semantics
- Migration plan: Add git version check at startup (`git --version`). Document minimum required git version. Have fallback for deprecated flags or behaviors.

**Zsh-Specific Syntax:**
- Risk: Code uses zsh-specific features (parameter expansion with `:t` suffix, `(( ... ))` arithmetic, `${...#...}` pattern removal) that don't work in bash or POSIX sh.
- Impact: Code cannot be ported to bash or used in non-zsh shells
- Migration plan: If portability becomes a goal, refactor to POSIX-compatible syntax or add shell detection with fallback implementation.

## Missing Critical Features

**No Dry-Run Mode:**
- Problem: `wt delete`, `wt eject`, and `wt new` make immediate changes with no `--dry-run` flag to preview what will happen
- Blocks: Users cannot safely test commands before executing; no way to validate naming/routing without doing the operation

**No Stash Recovery Help After Eject Failure:**
- Problem: If `_wt_eject` fails after creating stash, the stash is left dangling. Users must manually find and pop it.
- Blocks: Unclear state after failure; users may lose work

**No Batch Operations:**
- Problem: Deleting multiple worktrees requires running `wt delete` N times
- Blocks: Cleanup of many worktrees is tedious

**No Completion for Branch Names:**
- Problem: `wt new [flags] <name> [branch]` doesn't offer completion for existing branches
- Blocks: Users must type branch names manually

**No Config Validation:**
- Problem: .wtconfig file is not validated; invalid syntax or actions silently fail or produce warnings
- Blocks: Users may think file operations succeeded when they were skipped due to typos

## Test Coverage Gaps

**Error Path in Worktree Creation:**
- What's not tested: Behavior when `git worktree add` fails on second attempt (line 74). No test for when branch exists but worktree path doesn't.
- Files: `wt.zsh` (lines 72-78)
- Risk: Silent failure if both branch-creation and branch-checkout fail in unexpected ways
- Priority: High

**Stash Conflict Resolution:**
- What's not tested: Behavior when `git stash pop` encounters merge conflicts during `_wt_eject` (line 264). Current code silences stderr.
- Files: `wt.zsh` (lines 262-266)
- Risk: User loses visibility into conflicted stash state
- Priority: High

**Symlink Validation:**
- What's not tested: Whether symlinked files/directories persist after worktree is moved or filesystem is remounted
- Files: `wt.zsh` (lines 411-415, 437-441)
- Risk: Silent symlink breakage goes undetected
- Priority: Medium

**Ambiguous Name Resolution:**
- What's not tested: Behavior when multiple worktrees have overlapping suffix patterns (e.g., "main-foo" and "foo" both exist, and user runs `wt goto foo`)
- Files: `wt.zsh` (lines 454, 470, 498)
- Risk: Unpredictable worktree selection
- Priority: Medium

**Non-Standard Worktree Naming:**
- What's not tested: Behavior when `wt eject` is run in a worktree created outside the tool (e.g., via `git worktree add` directly with non-standard naming)
- Files: `wt.zsh` (lines 176-207)
- Risk: Fallback branch detection fails or selects wrong branch
- Priority: Medium

**Config + CLI Override Combinations:**
- What's not tested: All combinations of .wtconfig entries + CLI `--copy` and `--symlink` flags
- Files: `wt.zsh` (lines 368-444)
- Risk: Override precedence may not work as expected in edge cases
- Priority: Medium

---

*Concerns audit: 2026-02-07*
